package handlers

import (
	"context"
	"errors"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"

	api "github.com/skaletek/rule-engine-v2-poc/internal/api"
	"github.com/skaletek/rule-engine-v2-poc/internal/engine"
	"github.com/skaletek/rule-engine-v2-poc/internal/platform/db/sqlc"
	"github.com/skaletek/rule-engine-v2-poc/internal/rule"
)

func (h *Handler) ListRules(ctx context.Context, req api.ListRulesRequestObject) (api.ListRulesResponseObject, error) {
	limit := int32(20)
	offset := int32(0)
	if req.Params.Limit != nil {
		limit = int32(*req.Params.Limit)
	}
	if req.Params.Offset != nil {
		offset = int32(*req.Params.Offset)
	}

	var (
		rows []sqlc.Rule
		err  error
	)
	if req.Params.TemplateId != nil {
		rows, err = h.store.Queries.ListRulesByTemplate(ctx, sqlc.ListRulesByTemplateParams{
			TemplateID: pgtype.UUID{Bytes: *req.Params.TemplateId, Valid: true},
			Limit:      limit,
			Offset:     offset,
		})
	} else {
		rows, err = h.store.Queries.ListRules(ctx, sqlc.ListRulesParams{Limit: limit, Offset: offset})
	}
	if err != nil {
		return api.ListRules500JSONResponse{InternalErrorJSONResponse: api.InternalErrorJSONResponse{Message: err.Error()}}, nil
	}

	data := make([]api.Rule, len(rows))
	for i, r := range rows {
		data[i] = ruleToAPI(r)
	}
	return api.ListRules200JSONResponse{Data: data}, nil
}

func (h *Handler) CreateRule(ctx context.Context, req api.CreateRuleRequestObject) (api.CreateRuleResponseObject, error) {
	if errs := validateExpression(req.Body.Expression); len(errs) > 0 {
		return api.CreateRule422JSONResponse{
			UnprocessableEntityJSONResponse: api.UnprocessableEntityJSONResponse{
				Message: "invalid expression",
				Details: &errs,
			},
		}, nil
	}

	status := string(rule.StatusDraft)
	if req.Body.Status != nil {
		status = string(*req.Body.Status)
	}
	mode := string(rule.ModeLive)
	if req.Body.Mode != nil {
		mode = string(*req.Body.Mode)
	}
	priority := int32(10)
	if req.Body.Priority != nil {
		priority = int32(*req.Body.Priority)
	}
	message := ""
	if req.Body.Message != nil {
		message = *req.Body.Message
	}

	row, err := h.store.Queries.CreateRule(ctx, sqlc.CreateRuleParams{
		TemplateID: pgtype.UUID{Bytes: req.Body.TemplateId, Valid: true},
		Name:       req.Body.Name,
		Expression: req.Body.Expression,
		Severity:   string(req.Body.Severity),
		Message:    message,
		Priority:   priority,
		Status:     status,
		Mode:       mode,
	})
	if err != nil {
		return api.CreateRule500JSONResponse{InternalErrorJSONResponse: api.InternalErrorJSONResponse{Message: err.Error()}}, nil
	}
	return api.CreateRule201JSONResponse(ruleToAPI(row)), nil
}

func (h *Handler) GetRule(ctx context.Context, req api.GetRuleRequestObject) (api.GetRuleResponseObject, error) {
	row, err := h.store.Queries.GetRule(ctx, pgtype.UUID{Bytes: req.Id, Valid: true})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return api.GetRule404JSONResponse{NotFoundJSONResponse: api.NotFoundJSONResponse{Message: "rule not found"}}, nil
		}
		return api.GetRule500JSONResponse{InternalErrorJSONResponse: api.InternalErrorJSONResponse{Message: err.Error()}}, nil
	}
	return api.GetRule200JSONResponse(ruleToAPI(row)), nil
}

func (h *Handler) UpdateRule(ctx context.Context, req api.UpdateRuleRequestObject) (api.UpdateRuleResponseObject, error) {
	if errs := validateExpression(req.Body.Expression); len(errs) > 0 {
		return api.UpdateRule422JSONResponse{
			UnprocessableEntityJSONResponse: api.UnprocessableEntityJSONResponse{
				Message: "invalid expression",
				Details: &errs,
			},
		}, nil
	}

	existing, err := h.store.Queries.GetRule(ctx, pgtype.UUID{Bytes: req.Id, Valid: true})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return api.UpdateRule404JSONResponse{NotFoundJSONResponse: api.NotFoundJSONResponse{Message: "rule not found"}}, nil
		}
		return api.UpdateRule500JSONResponse{InternalErrorJSONResponse: api.InternalErrorJSONResponse{Message: err.Error()}}, nil
	}

	status := existing.Status
	if req.Body.Status != nil {
		status = string(*req.Body.Status)
	}
	mode := existing.Mode
	if req.Body.Mode != nil {
		mode = string(*req.Body.Mode)
	}
	priority := existing.Priority
	if req.Body.Priority != nil {
		priority = int32(*req.Body.Priority)
	}
	message := existing.Message
	if req.Body.Message != nil {
		message = *req.Body.Message
	}

	row, err := h.store.Queries.UpdateRule(ctx, sqlc.UpdateRuleParams{
		ID:         pgtype.UUID{Bytes: req.Id, Valid: true},
		Name:       req.Body.Name,
		Expression: req.Body.Expression,
		Severity:   string(req.Body.Severity),
		Message:    message,
		Priority:   priority,
		Status:     status,
		Mode:       mode,
	})
	if err != nil {
		return api.UpdateRule500JSONResponse{InternalErrorJSONResponse: api.InternalErrorJSONResponse{Message: err.Error()}}, nil
	}
	return api.UpdateRule200JSONResponse(ruleToAPI(row)), nil
}

func (h *Handler) DeleteRule(ctx context.Context, req api.DeleteRuleRequestObject) (api.DeleteRuleResponseObject, error) {
	err := h.store.Queries.DeleteRule(ctx, pgtype.UUID{Bytes: req.Id, Valid: true})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return api.DeleteRule404JSONResponse{NotFoundJSONResponse: api.NotFoundJSONResponse{Message: "rule not found"}}, nil
		}
		return api.DeleteRule500JSONResponse{InternalErrorJSONResponse: api.InternalErrorJSONResponse{Message: err.Error()}}, nil
	}
	return api.DeleteRule204Response{}, nil
}

// validateExpression compiles the expression with a probe rule and returns any errors.
func validateExpression(expression string) []string {
	probe := rule.Rule{
		ID:         "probe",
		Name:       "probe",
		Expression: expression,
		Status:     rule.StatusActive,
		Mode:       rule.ModeLive,
	}
	_, compileErrs := engine.CompileRules([]rule.Rule{probe})
	if len(compileErrs) == 0 {
		return nil
	}

	msgs := make([]string, len(compileErrs))
	for i, e := range compileErrs {
		msgs[i] = strings.TrimPrefix(e.Err.Error(), "probe (probe): ")
	}
	return msgs
}
