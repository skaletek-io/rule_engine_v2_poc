package handlers

import (
	"encoding/json"

	openapi_types "github.com/oapi-codegen/runtime/types"

	api "github.com/skaletek/rule-engine-v2-poc/internal/api"
	"github.com/skaletek/rule-engine-v2-poc/internal/platform/db/sqlc"
	"github.com/skaletek/rule-engine-v2-poc/internal/platform/db/store"
)

// Handler implements api.StrictServerInterface.
type Handler struct {
	store *store.Store
}

// New returns a Handler backed by store.
func New(s *store.Store) *Handler {
	return &Handler{store: s}
}

// ── Conversion helpers ────────────────────────────────────────────────────────

func toUUID(b [16]byte) openapi_types.UUID {
	return openapi_types.UUID(b)
}

func jsonToMap(data []byte) map[string]any {
	if len(data) == 0 {
		return map[string]any{}
	}
	var m map[string]any
	_ = json.Unmarshal(data, &m)
	return m
}

func mapToJSON(m map[string]any) []byte {
	if m == nil {
		return []byte("{}")
	}
	b, _ := json.Marshal(m)
	return b
}

func templateToAPI(t sqlc.Template) api.Template {
	return api.Template{
		Id:          toUUID(t.ID.Bytes),
		Slug:        t.Slug,
		Name:        t.Name,
		Description: t.Description,
		Schema:      jsonToMap(t.Schema),
		CreatedAt:   t.CreatedAt.Time,
		UpdatedAt:   t.UpdatedAt.Time,
	}
}

func eventToAPI(e sqlc.Event) api.Event {
	return api.Event{
		Id:          toUUID(e.ID.Bytes),
		TemplateId:  toUUID(e.TemplateID.Bytes),
		ExternalRef: e.ExternalRef,
		Payload:     jsonToMap(e.Payload),
		OccurredAt:  e.OccurredAt.Time,
		ReceivedAt:  e.ReceivedAt.Time,
	}
}

func ruleToAPI(r sqlc.Rule) api.Rule {
	return api.Rule{
		Id:         toUUID(r.ID.Bytes),
		TemplateId: toUUID(r.TemplateID.Bytes),
		Name:       r.Name,
		Expression: r.Expression,
		Severity:   api.RuleSeverity(r.Severity),
		Message:    r.Message,
		Priority:   int(r.Priority),
		Status:     api.RuleStatus(r.Status),
		Mode:       api.RuleMode(r.Mode),
		CreatedAt:  r.CreatedAt.Time,
		UpdatedAt:  r.UpdatedAt.Time,
	}
}
