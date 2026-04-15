package engine_test

import (
	"testing"
	"time"

	"github.com/skaletek/rule-engine-v2-poc/internal/engine"
	db "github.com/skaletek/rule-engine-v2-poc/internal/platform/db"
	"github.com/skaletek/rule-engine-v2-poc/internal/rule"
)

// BenchmarkCompileRules measures the cost of compiling all seed rules from
// scratch. This is a one-time startup cost; the result is intended to guide
// acceptable cold-start latency budgets.
func BenchmarkCompileRules(b *testing.B) {
	rules := rule.SeedRules()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		compiled, errs := engine.CompileRules(rules)
		if len(errs) != 0 {
			b.Fatalf("unexpected compile errors: %v", errs)
		}
		_ = compiled
	}
}

// BenchmarkEvaluateRules_AllRulesAllEvents measures steady-state evaluation
// throughput: 15 compiled rules evaluated against 12 seed events (180 rule×event
// evaluations per iteration after template filtering).
func BenchmarkEvaluateRules_AllRulesAllEvents(b *testing.B) {
	rules := rule.SeedRules()
	events := db.SeedEvents()

	compiled, errs := engine.CompileRules(rules)
	if len(errs) != 0 {
		b.Fatalf("compile errors: %v", errs)
	}

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		for _, ev := range events {
			_, _, err := engine.EvaluateRules(ev, compiled)
			if err != nil {
				b.Fatalf("eval error on %s: %v", ev.ID, err)
			}
		}
	}
}

// BenchmarkEvaluateRules_SingleRule_HotPath measures the minimum overhead of
// evaluating one rule against one event — the hot-path case.
func BenchmarkEvaluateRules_SingleRule_HotPath(b *testing.B) {
	r := rule.Rule{
		ID:         "bench_rule",
		Expression: `payment.amount > 10000 && sender.accountAge < 30`,
		Status:     rule.StatusActive,
		Mode:       rule.ModeLive,
	}
	ev := db.Event{
		ID:         "bench_event",
		TemplateID: "",
		OccurredAt: time.Now(),
		Payload: map[string]any{
			"payment": map[string]any{"amount": 45000.0},
			"sender":  map[string]any{"accountAge": 14},
		},
	}

	compiled, errs := engine.CompileRules([]rule.Rule{r})
	if len(errs) != 0 {
		b.Fatalf("compile error: %v", errs)
	}

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_, _, err := engine.EvaluateRules(ev, compiled)
		if err != nil {
			b.Fatalf("eval error: %v", err)
		}
	}
}

// BenchmarkBuildEvalContext measures the cost of building the evaluation context
// for a typical multi-role event payload.
func BenchmarkBuildEvalContext(b *testing.B) {
	ev := db.Event{
		ID:         "bench_ctx",
		TemplateID: "bank_wire_transfer",
		OccurredAt: time.Now(),
		Payload: map[string]any{
			"sender": map[string]any{
				"fullName":    "Alice",
				"nationality": "NG",
				"accountAge":  14,
			},
			"receiver": map[string]any{
				"fullName": "Bob",
				"country":  "SC",
			},
			"payment": map[string]any{
				"amount":   45000.0,
				"currency": "USD",
			},
			"channel": "online",
		},
	}

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		ctx := engine.BuildEvalContext(ev)
		_ = ctx
	}
}

// BenchmarkEvaluateRules_ComplexArithmetic benchmarks an expression that
// involves multiplication (representative of insurance-style rules).
func BenchmarkEvaluateRules_ComplexArithmetic(b *testing.B) {
	r := rule.Rule{
		ID:         "arith_bench",
		Expression: `claim.amountRequested >= policy.coverageLimit * 0.9`,
		Status:     rule.StatusActive,
		Mode:       rule.ModeLive,
	}
	ev := db.Event{
		ID:         "bench_arith",
		OccurredAt: time.Now(),
		Payload: map[string]any{
			"claim":  map[string]any{"amountRequested": 48000.0},
			"policy": map[string]any{"coverageLimit": 50000.0},
		},
	}

	compiled, errs := engine.CompileRules([]rule.Rule{r})
	if len(errs) != 0 {
		b.Fatalf("compile error: %v", errs)
	}

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_, _, err := engine.EvaluateRules(ev, compiled)
		if err != nil {
			b.Fatalf("eval error: %v", err)
		}
	}
}
