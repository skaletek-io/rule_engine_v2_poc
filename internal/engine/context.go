package engine

import (
	db "github.com/skaletek/rule-engine-v2-poc/internal/platform/db"
)

// BuildEvalContext builds the expr-lang evaluation context for an event.
// Payload roles are exposed by name (e.g. payment.amount, sender.country).
// Event metadata is exposed under the "event" key.
// No flattening is applied — expr-lang navigates map[string]any natively.
func BuildEvalContext(event db.Event) map[string]any {
	ctx := make(map[string]any, len(event.Payload)+1)
	for k, v := range event.Payload {
		ctx[k] = v
	}
	ctx["event"] = map[string]any{
		"id":         event.ID,
		"templateId": event.TemplateID,
		"occurredAt": event.OccurredAt,
	}
	return ctx
}
