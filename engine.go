package main

import (
	"fmt"

	"github.com/expr-lang/expr"
	"github.com/expr-lang/expr/vm"
)

type CompiledRule struct {
	Rule    Rule
	Program *vm.Program
}

func compileRules(rules []Rule) ([]CompiledRule, error) {
	var compiled []CompiledRule

	for _, rule := range rules {
		program, err := expr.Compile(rule.Expression)
		if err != nil {
			return nil, fmt.Errorf("failed to compile rule %s (%s): %w", rule.ID, rule.Name, err)
		}
		compiled = append(compiled, CompiledRule{Rule: rule, Program: program})
	}

	return compiled, nil
}

func evaluateRules(event Event, rules []CompiledRule) ([]Alert, error) {
	var alerts []Alert

	// flatten payload so expr can access fields directly
	ctx := flattenPayload(event.Payload)

	for _, cr := range rules {
		// skip rules not meant for this template
		if cr.Rule.TemplateID != "" && cr.Rule.TemplateID != event.TemplateID {
			continue
		}

		result, err := expr.Run(cr.Program, ctx)
		if err != nil {
			// non-fatal: rule may reference fields not present in this event
			fmt.Printf("  [skip] rule %s on event %s: %v\n", cr.Rule.ID, event.ID, err)
			continue
		}

		matched, ok := result.(bool)
		if !ok || !matched {
			continue
		}

		alerts = append(alerts, newAlert(event, cr.Rule))
	}

	return alerts, nil
}

// flattenPayload merges nested maps one level deep so expr can
// access sender.fullName, payment.amount etc via dot notation
func flattenPayload(payload map[string]any) map[string]any {
	flat := make(map[string]any)
	for k, v := range payload {
		flat[k] = v
		// also expose nested maps directly so top-level fields work too
		if nested, ok := v.(map[string]any); ok {
			for nk, nv := range nested {
				flat[nk] = nv
			}
		}
	}
	return flat
}
