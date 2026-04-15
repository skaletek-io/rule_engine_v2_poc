package engine_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/skaletek/rule-engine-v2-poc/internal/engine"
	db "github.com/skaletek/rule-engine-v2-poc/internal/platform/db"
	"github.com/skaletek/rule-engine-v2-poc/internal/rule"
)

// ── helpers ──────────────────────────────────────────────────────────────────

func mustCompile(t *testing.T, rules []rule.Rule) []engine.CompiledRule {
	t.Helper()
	compiled, errs := engine.CompileRules(rules)
	require.Empty(t, errs, "unexpected compile errors")
	return compiled
}

func firedIDs(results []engine.EvalResult) []string {
	var ids []string
	for _, r := range results {
		if r.Fired {
			ids = append(ids, r.RuleID)
		}
	}
	return ids
}

func seedRule(id string) rule.Rule {
	for _, r := range rule.SeedRules() {
		if r.ID == id {
			return r
		}
	}
	panic("unknown seed rule: " + id)
}

func seedEvent(id string) db.Event {
	for _, e := range db.SeedEvents() {
		if e.ID == id {
			return e
		}
	}
	panic("unknown seed event: " + id)
}

// ── 1. CompileRules ───────────────────────────────────────────────────────────

func TestCompileRules_AllSeedRulesCompile(t *testing.T) {
	rules := rule.SeedRules()
	compiled, errs := engine.CompileRules(rules)
	require.Empty(t, errs)
	require.Len(t, compiled, len(rules))
}

func TestCompileRules_SkipsDraftAndDisabled(t *testing.T) {
	rules := []rule.Rule{
		{ID: "active", Expression: `x > 0`, Status: rule.StatusActive, Mode: rule.ModeLive},
		{ID: "draft", Expression: `x > 0`, Status: rule.StatusDraft, Mode: rule.ModeLive},
		{ID: "disabled", Expression: `x > 0`, Status: rule.StatusDisabled, Mode: rule.ModeLive},
	}
	compiled, errs := engine.CompileRules(rules)
	require.Empty(t, errs)
	require.Len(t, compiled, 1)
	require.Equal(t, "active", compiled[0].Rule.ID)
}

func TestCompileRules_CollectsAllErrors(t *testing.T) {
	rules := []rule.Rule{
		{ID: "bad_1", Expression: `this is not valid ###`, Status: rule.StatusActive, Mode: rule.ModeLive},
		{ID: "bad_2", Expression: `another bad ##`, Status: rule.StatusActive, Mode: rule.ModeLive},
		{ID: "good", Expression: `x > 0`, Status: rule.StatusActive, Mode: rule.ModeLive},
	}
	compiled, errs := engine.CompileRules(rules)
	require.Len(t, errs, 2)
	require.Len(t, compiled, 1)
	require.Equal(t, "good", compiled[0].Rule.ID)
}

func TestCompileRules_NonBoolExpressionRejected(t *testing.T) {
	rules := []rule.Rule{
		{ID: "non_bool", Expression: `1 + 2`, Status: rule.StatusActive, Mode: rule.ModeLive},
	}
	_, errs := engine.CompileRules(rules)
	require.NotEmpty(t, errs)
}

func TestCompileRules_PrioritySorting(t *testing.T) {
	rules := []rule.Rule{
		{ID: "r30", Expression: `true`, Status: rule.StatusActive, Mode: rule.ModeLive, Priority: 30},
		{ID: "r5", Expression: `true`, Status: rule.StatusActive, Mode: rule.ModeLive, Priority: 5},
		{ID: "r20", Expression: `true`, Status: rule.StatusActive, Mode: rule.ModeLive, Priority: 20},
		{ID: "r1", Expression: `true`, Status: rule.StatusActive, Mode: rule.ModeLive, Priority: 1},
	}
	compiled, errs := engine.CompileRules(rules)
	require.Empty(t, errs)
	require.Equal(t, []string{"r1", "r5", "r20", "r30"}, []string{
		compiled[0].Rule.ID, compiled[1].Rule.ID, compiled[2].Rule.ID, compiled[3].Rule.ID,
	})
}

// ── 2. BuildEvalContext ───────────────────────────────────────────────────────

func TestBuildEvalContext_ExposesRoleFields(t *testing.T) {
	ev := db.Event{
		ID:         "evt_test",
		TemplateID: "test_template",
		OccurredAt: time.Now(),
		Payload: map[string]any{
			"payment": map[string]any{"amount": 5000.0},
			"sender":  map[string]any{"name": "Alice", "accountAge": 10},
		},
	}
	ctx := engine.BuildEvalContext(ev)

	payment := ctx["payment"].(map[string]any)
	require.Equal(t, 5000.0, payment["amount"])

	sender := ctx["sender"].(map[string]any)
	require.Equal(t, "Alice", sender["name"])
}

func TestBuildEvalContext_ExposesEventMetadata(t *testing.T) {
	now := time.Now()
	ev := db.Event{ID: "evt_meta", TemplateID: "tmpl_meta", OccurredAt: now, Payload: map[string]any{}}
	ctx := engine.BuildEvalContext(ev)

	meta := ctx["event"].(map[string]any)
	require.Equal(t, "evt_meta", meta["id"])
	require.Equal(t, "tmpl_meta", meta["templateId"])
	require.Equal(t, now, meta["occurredAt"])
}

func TestBuildEvalContext_NoKeyCollisions(t *testing.T) {
	ev := db.Event{
		ID:         "evt_collision",
		TemplateID: "bank_wire_transfer",
		OccurredAt: time.Now(),
		Payload: map[string]any{
			"sender":   map[string]any{"fullName": "Alice"},
			"receiver": map[string]any{"fullName": "Bob"},
		},
	}
	ctx := engine.BuildEvalContext(ev)
	require.Equal(t, "Alice", ctx["sender"].(map[string]any)["fullName"])
	require.Equal(t, "Bob", ctx["receiver"].(map[string]any)["fullName"])
}

// ── 3. Correctness: per-template rule × event pairs ──────────────────────────

func TestCorrectness_BankingRules(t *testing.T) {
	cases := []struct {
		name    string
		ruleID  string
		eventID string
		want    bool
	}{
		{"bank_001 fires on high-value new-account wire", "rule_bank_001", "evt_bank_001", true},
		{"bank_001 does not fire on low-value established-account wire", "rule_bank_001", "evt_bank_002", false},
		{"bank_002 fires on SC destination", "rule_bank_002", "evt_bank_001", true},
		{"bank_002 does not fire on NG destination", "rule_bank_002", "evt_bank_002", false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			compiled := mustCompile(t, []rule.Rule{seedRule(tc.ruleID)})
			_, results, err := engine.EvaluateRules(seedEvent(tc.eventID), compiled)
			require.NoError(t, err)
			require.NotEmpty(t, results, "rule was skipped (template mismatch?)")
			require.Equal(t, tc.want, results[0].Fired, "err=%v", results[0].Err)
		})
	}
}

func TestCorrectness_FintechRules(t *testing.T) {
	cases := []struct {
		name    string
		ruleID  string
		eventID string
		want    bool
	}{
		{"fin_001 fires on unverified large tx", "rule_fin_001", "evt_fin_001", true},
		{"fin_001 does not fire on verified small tx", "rule_fin_001", "evt_fin_002", false},
		{"fin_002 fires on VPN+crypto merchant", "rule_fin_002", "evt_fin_001", true},
		{"fin_002 does not fire on normal tx", "rule_fin_002", "evt_fin_002", false},
		{"fin_003 fires on new account high value", "rule_fin_003", "evt_fin_001", true},
		{"fin_003 does not fire on old account low value", "rule_fin_003", "evt_fin_002", false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			compiled := mustCompile(t, []rule.Rule{seedRule(tc.ruleID)})
			_, results, err := engine.EvaluateRules(seedEvent(tc.eventID), compiled)
			require.NoError(t, err)
			require.NotEmpty(t, results)
			require.Equal(t, tc.want, results[0].Fired, "err=%v", results[0].Err)
		})
	}
}

func TestCorrectness_CryptoRules(t *testing.T) {
	cases := []struct {
		name    string
		ruleID  string
		eventID string
		want    bool
	}{
		{"crypto_001 fires on blacklisted wallet", "rule_crypto_001", "evt_crypto_001", true},
		{"crypto_001 does not fire on clean wallet", "rule_crypto_001", "evt_crypto_002", false},
		{"crypto_002 fires on high-risk large withdrawal", "rule_crypto_002", "evt_crypto_001", true},
		{"crypto_002 does not fire on low-risk small withdrawal", "rule_crypto_002", "evt_crypto_002", false},
		{"crypto_003 fires on unverified large withdrawal", "rule_crypto_003", "evt_crypto_001", true},
		{"crypto_003 does not fire on verified small withdrawal", "rule_crypto_003", "evt_crypto_002", false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			compiled := mustCompile(t, []rule.Rule{seedRule(tc.ruleID)})
			_, results, err := engine.EvaluateRules(seedEvent(tc.eventID), compiled)
			require.NoError(t, err)
			require.NotEmpty(t, results)
			require.Equal(t, tc.want, results[0].Fired, "err=%v", results[0].Err)
		})
	}
}

func TestCorrectness_HotelRules(t *testing.T) {
	cases := []struct {
		name    string
		ruleID  string
		eventID string
		want    bool
	}{
		{"hotel_001 fires on bulk same-day crypto booking", "rule_hotel_001", "evt_hotel_001", true},
		{"hotel_001 does not fire on normal booking", "rule_hotel_001", "evt_hotel_002", false},
		{"hotel_002 fires on high-value sanctioned nationality", "rule_hotel_002", "evt_hotel_001", true},
		{"hotel_002 does not fire on normal booking", "rule_hotel_002", "evt_hotel_002", false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			compiled := mustCompile(t, []rule.Rule{seedRule(tc.ruleID)})
			_, results, err := engine.EvaluateRules(seedEvent(tc.eventID), compiled)
			require.NoError(t, err)
			require.NotEmpty(t, results)
			require.Equal(t, tc.want, results[0].Fired, "err=%v", results[0].Err)
		})
	}
}

func TestCorrectness_InsuranceRules(t *testing.T) {
	cases := []struct {
		name    string
		ruleID  string
		eventID string
		want    bool
	}{
		{"ins_001 fires on new policy large claim", "rule_ins_001", "evt_ins_001", true},
		{"ins_001 does not fire on old policy small claim", "rule_ins_001", "evt_ins_002", false},
		{"ins_002 fires on repeat claimant high value", "rule_ins_002", "evt_ins_001", true},
		{"ins_002 does not fire on first-time small claim", "rule_ins_002", "evt_ins_002", false},
		// 48000 >= 50000*0.9=45000 → true
		{"ins_003 fires on claim near coverage limit", "rule_ins_003", "evt_ins_001", true},
		// 3200 >= 100000*0.9=90000 → false
		{"ins_003 does not fire when claim is far below limit", "rule_ins_003", "evt_ins_002", false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			compiled := mustCompile(t, []rule.Rule{seedRule(tc.ruleID)})
			_, results, err := engine.EvaluateRules(seedEvent(tc.eventID), compiled)
			require.NoError(t, err)
			require.NotEmpty(t, results)
			require.Equal(t, tc.want, results[0].Fired, "err=%v", results[0].Err)
		})
	}
}

func TestCorrectness_EcommerceRules(t *testing.T) {
	cases := []struct {
		name    string
		ruleID  string
		eventID string
		want    bool
	}{
		{"ecom_001 fires on new account bulk electronics", "rule_ecom_001", "evt_ecom_001", true},
		{"ecom_001 does not fire on regular order", "rule_ecom_001", "evt_ecom_002", false},
		{"ecom_002 fires on mismatch+new card", "rule_ecom_002", "evt_ecom_001", true},
		{"ecom_002 does not fire on matching address+existing card", "rule_ecom_002", "evt_ecom_002", false},
		{"ecom_003 fires on high-value new account order", "rule_ecom_003", "evt_ecom_001", true},
		{"ecom_003 does not fire on regular buyer", "rule_ecom_003", "evt_ecom_002", false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			compiled := mustCompile(t, []rule.Rule{seedRule(tc.ruleID)})
			_, results, err := engine.EvaluateRules(seedEvent(tc.eventID), compiled)
			require.NoError(t, err)
			require.NotEmpty(t, results)
			require.Equal(t, tc.want, results[0].Fired, "err=%v", results[0].Err)
		})
	}
}

// ── 4. Template isolation ─────────────────────────────────────────────────────

// A rule must never fire on an event from a different template.
func TestTemplateScopeIsolation(t *testing.T) {
	allRules := rule.SeedRules()
	allEvents := db.SeedEvents()
	compiled := mustCompile(t, allRules)

	ruleByID := make(map[string]rule.Rule, len(allRules))
	for _, r := range allRules {
		ruleByID[r.ID] = r
	}

	for _, ev := range allEvents {
		_, results, err := engine.EvaluateRules(ev, compiled)
		require.NoError(t, err, "event %s", ev.ID)
		for _, res := range results {
			if res.Fired {
				r := ruleByID[res.RuleID]
				require.Truef(t,
					r.TemplateID == "" || r.TemplateID == ev.TemplateID,
					"rule %s (template=%s) fired on event %s (template=%s)",
					res.RuleID, r.TemplateID, ev.ID, ev.TemplateID,
				)
			}
		}
	}
}

// ── 5. Full accuracy matrix ───────────────────────────────────────────────────

func TestFullMatrix(t *testing.T) {
	shouldFire := map[string]bool{
		"rule_bank_001|evt_bank_001":     true,
		"rule_bank_002|evt_bank_001":     true,
		"rule_fin_001|evt_fin_001":       true,
		"rule_fin_002|evt_fin_001":       true,
		"rule_fin_003|evt_fin_001":       true,
		"rule_crypto_001|evt_crypto_001": true,
		"rule_crypto_002|evt_crypto_001": true,
		"rule_crypto_003|evt_crypto_001": true,
		"rule_hotel_001|evt_hotel_001":   true,
		"rule_hotel_002|evt_hotel_001":   true,
		"rule_ins_001|evt_ins_001":       true,
		"rule_ins_002|evt_ins_001":       true,
		"rule_ins_003|evt_ins_001":       true,
		"rule_ecom_001|evt_ecom_001":     true,
		"rule_ecom_002|evt_ecom_001":     true,
		"rule_ecom_003|evt_ecom_001":     true,
	}

	compiled := mustCompile(t, rule.SeedRules())

	for _, ev := range db.SeedEvents() {
		_, results, err := engine.EvaluateRules(ev, compiled)
		require.NoError(t, err, "event %s", ev.ID)
		for _, res := range results {
			if res.Err != nil {
				continue
			}
			key := res.RuleID + "|" + ev.ID
			require.Equalf(t, shouldFire[key], res.Fired,
				"rule=%s event=%s", res.RuleID, ev.ID)
		}
	}
}

// ── 6. Edge cases ─────────────────────────────────────────────────────────────

func TestEdgeCase_IntVsFloat64Comparison(t *testing.T) {
	r := rule.Rule{
		ID:         "int_compare",
		Expression: `tx.mcc == 6051 && tx.count > 3`,
		Status:     rule.StatusActive,
		Mode:       rule.ModeLive,
	}
	ev := db.Event{
		ID:         "evt_int",
		OccurredAt: time.Now(),
		Payload:    map[string]any{"tx": map[string]any{"mcc": 6051, "count": 10}},
	}
	compiled := mustCompile(t, []rule.Rule{r})
	_, results, err := engine.EvaluateRules(ev, compiled)
	require.NoError(t, err)
	require.True(t, results[0].Fired, "err=%v", results[0].Err)
}

func TestEdgeCase_BoolComparison(t *testing.T) {
	r := rule.Rule{
		ID:         "bool_compare",
		Expression: `user.verified == false && user.active == true`,
		Status:     rule.StatusActive,
		Mode:       rule.ModeLive,
	}
	ev := db.Event{
		ID:         "evt_bool",
		OccurredAt: time.Now(),
		Payload:    map[string]any{"user": map[string]any{"verified": false, "active": true}},
	}
	compiled := mustCompile(t, []rule.Rule{r})
	_, results, err := engine.EvaluateRules(ev, compiled)
	require.NoError(t, err)
	require.True(t, results[0].Fired, "err=%v", results[0].Err)
}

func TestEdgeCase_ArithmeticExpression(t *testing.T) {
	// 48000 >= 50000*0.9=45000 → true
	compiled := mustCompile(t, []rule.Rule{seedRule("rule_ins_003")})
	_, results, err := engine.EvaluateRules(seedEvent("evt_ins_001"), compiled)
	require.NoError(t, err)
	require.True(t, results[0].Fired, "err=%v", results[0].Err)
}

func TestEdgeCase_MissingFieldIsNonFatal(t *testing.T) {
	r := rule.Rule{
		ID:         "missing_field",
		Expression: `nonexistent.field > 100`,
		Status:     rule.StatusActive,
		Mode:       rule.ModeLive,
	}
	ev := db.Event{ID: "evt_empty", OccurredAt: time.Now(), Payload: map[string]any{}}
	compiled := mustCompile(t, []rule.Rule{r})
	alerts, results, err := engine.EvaluateRules(ev, compiled)
	require.NoError(t, err)
	require.Empty(t, alerts)
	require.Error(t, results[0].Err)
	require.False(t, results[0].Fired)
}

func TestEdgeCase_ShadowMode(t *testing.T) {
	r := rule.Rule{
		ID:         "shadow_rule",
		Expression: `amount > 0`,
		Status:     rule.StatusActive,
		Mode:       rule.ModeShadow,
	}
	ev := db.Event{ID: "evt_shadow", OccurredAt: time.Now(), Payload: map[string]any{"amount": 500.0}}
	compiled := mustCompile(t, []rule.Rule{r})
	alerts, results, err := engine.EvaluateRules(ev, compiled)
	require.NoError(t, err)
	require.Empty(t, alerts, "shadow mode must produce no alerts")
	require.True(t, results[0].Fired)
	require.True(t, results[0].Shadow)
}

func TestEdgeCase_MultipleRulesOnSameEvent(t *testing.T) {
	compiled := mustCompile(t, []rule.Rule{seedRule("rule_bank_001"), seedRule("rule_bank_002")})
	alerts, results, err := engine.EvaluateRules(seedEvent("evt_bank_001"), compiled)
	require.NoError(t, err)
	require.ElementsMatch(t, []string{"rule_bank_001", "rule_bank_002"}, firedIDs(results))
	require.Len(t, alerts, 2)
}

func TestEdgeCase_StringContains(t *testing.T) {
	r := rule.Rule{
		ID:         "string_contains",
		Expression: `email contains "@tempmail"`,
		Status:     rule.StatusActive,
		Mode:       rule.ModeLive,
	}
	ev := db.Event{ID: "evt_str", OccurredAt: time.Now(), Payload: map[string]any{"email": "spoofed@tempmail.io"}}
	compiled := mustCompile(t, []rule.Rule{r})
	_, results, err := engine.EvaluateRules(ev, compiled)
	require.NoError(t, err)
	require.True(t, results[0].Fired, "err=%v", results[0].Err)
}

func TestEdgeCase_StringMatches(t *testing.T) {
	r := rule.Rule{
		ID:         "regex_match",
		Expression: `email matches "^[a-z]+@tempmail\\.io$"`,
		Status:     rule.StatusActive,
		Mode:       rule.ModeLive,
	}
	ev := db.Event{ID: "evt_regex", OccurredAt: time.Now(), Payload: map[string]any{"email": "spoofed@tempmail.io"}}
	compiled := mustCompile(t, []rule.Rule{r})
	_, results, err := engine.EvaluateRules(ev, compiled)
	require.NoError(t, err)
	require.True(t, results[0].Fired, "err=%v", results[0].Err)
}

func TestEdgeCase_InOperator(t *testing.T) {
	r := rule.Rule{
		ID:         "in_list",
		Expression: `country in ["SC", "KY", "VU"]`,
		Status:     rule.StatusActive,
		Mode:       rule.ModeLive,
	}
	ev := db.Event{ID: "evt_in", OccurredAt: time.Now(), Payload: map[string]any{"country": "SC"}}
	compiled := mustCompile(t, []rule.Rule{r})
	_, results, err := engine.EvaluateRules(ev, compiled)
	require.NoError(t, err)
	require.True(t, results[0].Fired, "err=%v", results[0].Err)
}

func TestEdgeCase_NegationOperator(t *testing.T) {
	r := rule.Rule{
		ID:         "negation",
		Expression: `!(verified) && amount > 100`,
		Status:     rule.StatusActive,
		Mode:       rule.ModeLive,
	}
	ev := db.Event{
		ID:         "evt_neg",
		OccurredAt: time.Now(),
		Payload:    map[string]any{"verified": false, "amount": 500.0},
	}
	compiled := mustCompile(t, []rule.Rule{r})
	_, results, err := engine.EvaluateRules(ev, compiled)
	require.NoError(t, err)
	require.True(t, results[0].Fired, "err=%v", results[0].Err)
}

func TestEdgeCase_WordFormLogicOperators(t *testing.T) {
	r := rule.Rule{
		ID:         "word_logic",
		Expression: `verified == false and amount > 100 or risk == "high"`,
		Status:     rule.StatusActive,
		Mode:       rule.ModeLive,
	}
	ev := db.Event{
		ID:         "evt_word",
		OccurredAt: time.Now(),
		Payload:    map[string]any{"verified": false, "amount": 500.0, "risk": "low"},
	}
	compiled := mustCompile(t, []rule.Rule{r})
	_, results, err := engine.EvaluateRules(ev, compiled)
	require.NoError(t, err)
	require.True(t, results[0].Fired, "err=%v", results[0].Err)
}

func TestEdgeCase_EvalResultDurationIsPopulated(t *testing.T) {
	r := rule.Rule{ID: "duration_check", Expression: `x > 0`, Status: rule.StatusActive, Mode: rule.ModeLive}
	ev := db.Event{ID: "evt_dur", OccurredAt: time.Now(), Payload: map[string]any{"x": 1}}
	compiled := mustCompile(t, []rule.Rule{r})
	_, results, _ := engine.EvaluateRules(ev, compiled)
	require.NotZero(t, results[0].Duration)
}

func TestEdgeCase_EventMetadataInExpression(t *testing.T) {
	r := rule.Rule{
		ID:         "event_meta",
		Expression: `event.id == "evt_meta_test" && event.templateId == "my_template"`,
		Status:     rule.StatusActive,
		Mode:       rule.ModeLive,
	}
	ev := db.Event{ID: "evt_meta_test", TemplateID: "my_template", OccurredAt: time.Now(), Payload: map[string]any{}}
	compiled := mustCompile(t, []rule.Rule{r})
	_, results, err := engine.EvaluateRules(ev, compiled)
	require.NoError(t, err)
	require.True(t, results[0].Fired, "err=%v", results[0].Err)
}

// ── 7. CompileError ───────────────────────────────────────────────────────────

func TestCompileError_ErrorString(t *testing.T) {
	ce := engine.CompileError{RuleID: "rule_x", RuleName: "Test Rule", Err: fmt.Errorf("parse error")}
	require.Contains(t, ce.Error(), "rule_x")
}
