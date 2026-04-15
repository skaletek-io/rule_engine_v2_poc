// Package engine compiles and evaluates rule expressions using expr-lang.
//
// Supported grammar (expr-lang syntax):
//   - Logical:    && || !  (or: and or not)
//   - Comparison: == != < <= > >=
//   - String:     contains, matches (regex), startsWith, endsWith
//   - Membership: in [...]
//   - Arithmetic: + - * / %
//   - Field paths: role.field.subfield
//   - Event meta:  event.id, event.templateId, event.occurredAt
package engine

import (
	"fmt"
	"log"
	"sort"
	"time"

	"github.com/expr-lang/expr"
	"github.com/expr-lang/expr/vm"

	alert "github.com/skaletek/rule-engine-v2-poc/internal/alert"
	db "github.com/skaletek/rule-engine-v2-poc/internal/platform/db"
	rule "github.com/skaletek/rule-engine-v2-poc/internal/rule"
)

// DefaultEvalTimeout is the per-rule evaluation budget.
const DefaultEvalTimeout = 500 * time.Millisecond

// CompileError captures a compile-time failure for a single rule.
type CompileError struct {
	RuleID   string
	RuleName string
	Err      error
}

func (e CompileError) Error() string {
	return fmt.Sprintf("rule %s (%s): %v", e.RuleID, e.RuleName, e.Err)
}

// CompiledRule pairs a Rule definition with its compiled expr-lang program.
type CompiledRule struct {
	Rule    rule.Rule
	Program *vm.Program
}

// EvalResult records the outcome of evaluating one rule against one event.
type EvalResult struct {
	RuleID   string
	RuleName string
	Fired    bool
	Shadow   bool
	Duration time.Duration
	Err      error
}

// CompileRules compiles all active rules, collecting errors without failing
// fast. Rules are sorted by Priority ascending after compilation.
func CompileRules(rules []rule.Rule) ([]CompiledRule, []CompileError) {
	compiled := make([]CompiledRule, 0, len(rules))
	var errs []CompileError

	for _, r := range rules {
		if r.Status != rule.StatusActive {
			continue
		}
		// expr.AsBool enforces boolean return type. We skip expr.Env because
		// rules reference template-specific field paths that vary per template;
		// field existence is checked at evaluation time.
		program, err := expr.Compile(r.Expression, expr.AsBool())
		if err != nil {
			errs = append(errs, CompileError{RuleID: r.ID, RuleName: r.Name, Err: err})
			continue
		}
		compiled = append(compiled, CompiledRule{Rule: r, Program: program})
	}

	sort.Slice(compiled, func(i, j int) bool {
		return compiled[i].Rule.Priority < compiled[j].Rule.Priority
	})

	return compiled, errs
}

// EvaluateRules runs compiled rules against a single event. Rules whose
// TemplateID doesn't match the event are skipped. Shadow-mode rules are
// evaluated but produce no alerts. Per-rule errors are non-fatal.
func EvaluateRules(event db.Event, rules []CompiledRule) ([]alert.Alert, []EvalResult, error) {
	ctx := BuildEvalContext(event)

	var alerts []alert.Alert
	results := make([]EvalResult, 0, len(rules))

	for _, cr := range rules {
		if cr.Rule.TemplateID != "" && cr.Rule.TemplateID != event.TemplateID {
			continue
		}

		res := evalOne(cr, ctx)
		results = append(results, res)

		if res.Err != nil {
			log.Printf("[engine] skip rule %s on event %s: %v", cr.Rule.ID, event.ID, res.Err)
			continue
		}
		if !res.Fired {
			continue
		}
		if res.Shadow {
			log.Printf("[engine] shadow match rule %s on event %s", cr.Rule.ID, event.ID)
			continue
		}

		alerts = append(alerts, alert.NewAlert(event, cr.Rule))
	}

	return alerts, results, nil
}

// evalOne evaluates a single rule within DefaultEvalTimeout.
func evalOne(cr CompiledRule, ctx map[string]any) EvalResult {
	res := EvalResult{
		RuleID:   cr.Rule.ID,
		RuleName: cr.Rule.Name,
		Shadow:   cr.Rule.Mode == rule.ModeShadow,
	}

	type outcome struct {
		val any
		err error
	}
	ch := make(chan outcome, 1)

	start := time.Now()
	go func() {
		val, err := expr.Run(cr.Program, ctx)
		ch <- outcome{val, err}
	}()

	select {
	case out := <-ch:
		res.Duration = time.Since(start)
		if out.err != nil {
			res.Err = out.err
			return res
		}
		matched, ok := out.val.(bool)
		if !ok {
			res.Err = fmt.Errorf("expression returned non-bool: %T", out.val)
			return res
		}
		res.Fired = matched

	case <-time.After(DefaultEvalTimeout):
		res.Duration = DefaultEvalTimeout
		res.Err = fmt.Errorf("rule evaluation timed out after %s", DefaultEvalTimeout)
	}

	return res
}
