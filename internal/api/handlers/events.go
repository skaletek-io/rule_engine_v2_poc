package handlers

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	openapi_types "github.com/oapi-codegen/runtime/types"

	api "github.com/skaletek/rule-engine-v2-poc/internal/api"
	"github.com/skaletek/rule-engine-v2-poc/internal/engine"
	platformdb "github.com/skaletek/rule-engine-v2-poc/internal/platform/db"
	"github.com/skaletek/rule-engine-v2-poc/internal/platform/db/sqlc"
	"github.com/skaletek/rule-engine-v2-poc/internal/rule"
)

func (h *Handler) ListEvents(ctx context.Context, req api.ListEventsRequestObject) (api.ListEventsResponseObject, error) {
	limit := int32(20)
	offset := int32(0)
	if req.Params.Limit != nil {
		limit = int32(*req.Params.Limit)
	}
	if req.Params.Offset != nil {
		offset = int32(*req.Params.Offset)
	}

	var (
		rows []sqlc.Event
		err  error
	)
	if req.Params.TemplateId != nil {
		rows, err = h.store.ListEventsByTemplate(ctx, sqlc.ListEventsByTemplateParams{
			TemplateID: pgtype.UUID{Bytes: *req.Params.TemplateId, Valid: true},
			Limit:      limit,
			Offset:     offset,
		})
	} else {
		rows, err = h.store.ListEvents(ctx, sqlc.ListEventsParams{Limit: limit, Offset: offset})
	}
	if err != nil {
		return api.ListEvents500JSONResponse{InternalErrorJSONResponse: api.InternalErrorJSONResponse{Message: err.Error()}}, nil
	}

	data := make([]api.Event, len(rows))
	for i, r := range rows {
		data[i] = eventToAPI(r)
	}
	return api.ListEvents200JSONResponse{Data: data}, nil
}

func (h *Handler) CreateEvent(ctx context.Context, req api.CreateEventRequestObject) (api.CreateEventResponseObject, error) {
	dbEvent, err := h.store.Queries.CreateEvent(ctx, sqlc.CreateEventParams{
		TemplateID:  pgtype.UUID{Bytes: req.Body.TemplateId, Valid: true},
		ExternalRef: req.Body.ExternalRef,
		Payload:     mapToJSON(req.Body.Payload),
		OccurredAt:  pgtype.Timestamptz{Time: req.Body.OccurredAt, Valid: true},
	})
	if err != nil {
		return api.CreateEvent500JSONResponse{InternalErrorJSONResponse: api.InternalErrorJSONResponse{Message: err.Error()}}, nil
	}

	firedAlerts, err := h.evaluateRules(ctx, dbEvent)
	if err != nil {
		fmt.Printf("[events] rule evaluation error: %v\n", err)
	}

	return api.CreateEvent201JSONResponse{
		Event:  eventToAPI(dbEvent),
		Alerts: firedAlerts,
	}, nil
}

func (h *Handler) GetEvent(ctx context.Context, req api.GetEventRequestObject) (api.GetEventResponseObject, error) {
	row, err := h.store.Queries.GetEvent(ctx, pgtype.UUID{Bytes: req.Id, Valid: true})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return api.GetEvent404JSONResponse{NotFoundJSONResponse: api.NotFoundJSONResponse{Message: "event not found"}}, nil
		}
		return api.GetEvent500JSONResponse{InternalErrorJSONResponse: api.InternalErrorJSONResponse{Message: err.Error()}}, nil
	}
	return api.GetEvent200JSONResponse(eventToAPI(row)), nil
}

func (h *Handler) UpdateEvent(ctx context.Context, req api.UpdateEventRequestObject) (api.UpdateEventResponseObject, error) {
	row, err := h.store.Queries.UpdateEvent(ctx, sqlc.UpdateEventParams{
		ID:          pgtype.UUID{Bytes: req.Id, Valid: true},
		ExternalRef: req.Body.ExternalRef,
		Payload:     mapToJSON(req.Body.Payload),
		OccurredAt:  pgtype.Timestamptz{Time: req.Body.OccurredAt, Valid: true},
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return api.UpdateEvent404JSONResponse{NotFoundJSONResponse: api.NotFoundJSONResponse{Message: "event not found"}}, nil
		}
		return api.UpdateEvent500JSONResponse{InternalErrorJSONResponse: api.InternalErrorJSONResponse{Message: err.Error()}}, nil
	}
	return api.UpdateEvent200JSONResponse(eventToAPI(row)), nil
}

func (h *Handler) DeleteEvent(ctx context.Context, req api.DeleteEventRequestObject) (api.DeleteEventResponseObject, error) {
	err := h.store.Queries.DeleteEvent(ctx, pgtype.UUID{Bytes: req.Id, Valid: true})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return api.DeleteEvent404JSONResponse{NotFoundJSONResponse: api.NotFoundJSONResponse{Message: "event not found"}}, nil
		}
		return api.DeleteEvent500JSONResponse{InternalErrorJSONResponse: api.InternalErrorJSONResponse{Message: err.Error()}}, nil
	}
	return api.DeleteEvent204Response{}, nil
}

// evaluateRules fetches active rules for the event's template, compiles and runs them,
// returning fired alerts.
func (h *Handler) evaluateRules(ctx context.Context, ev sqlc.Event) ([]api.FiredAlert, error) {
	dbRules, err := h.store.Queries.ListActiveRulesByTemplate(ctx, ev.TemplateID)
	if err != nil {
		return nil, err
	}

	// Build a lookup from string rule ID → sqlc.Rule for fired-alert mapping.
	ruleByID := make(map[string]sqlc.Rule, len(dbRules))
	engineRules := make([]rule.Rule, len(dbRules))
	for i, r := range dbRules {
		sid := uuid.UUID(r.ID.Bytes).String()
		ruleByID[sid] = r
		engineRules[i] = rule.Rule{
			ID:         sid,
			Name:       r.Name,
			TemplateID: uuid.UUID(r.TemplateID.Bytes).String(),
			Expression: r.Expression,
			Severity:   r.Severity,
			Message:    r.Message,
			Priority:   int(r.Priority),
			Status:     r.Status,
			Mode:       r.Mode,
		}
	}

	compiled, _ := engine.CompileRules(engineRules)
	if len(compiled) == 0 {
		return []api.FiredAlert{}, nil
	}

	platformEvent := platformdb.Event{
		ID:         uuid.UUID(ev.ID.Bytes).String(),
		TemplateID: uuid.UUID(ev.TemplateID.Bytes).String(),
		OccurredAt: ev.OccurredAt.Time,
		Payload:    jsonToMap(ev.Payload),
	}

	alerts, _, _ := engine.EvaluateRules(platformEvent, compiled)

	apiAlerts := make([]api.FiredAlert, 0, len(alerts))
	for _, a := range alerts {
		dbRule, ok := ruleByID[a.RuleID]
		if !ok {
			continue
		}
		apiAlerts = append(apiAlerts, api.FiredAlert{
			RuleId:   openapi_types.UUID(dbRule.ID.Bytes),
			RuleName: a.RuleName,
			Severity: a.Severity,
			Message:  a.Message,
		})
	}

	return apiAlerts, nil
}
