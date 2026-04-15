package handlers

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"

	api "github.com/skaletek/rule-engine-v2-poc/internal/api"
	"github.com/skaletek/rule-engine-v2-poc/internal/platform/db/sqlc"
)

func (h *Handler) ListTemplates(ctx context.Context, req api.ListTemplatesRequestObject) (api.ListTemplatesResponseObject, error) {
	limit := int32(20)
	offset := int32(0)
	if req.Params.Limit != nil {
		limit = int32(*req.Params.Limit)
	}
	if req.Params.Offset != nil {
		offset = int32(*req.Params.Offset)
	}

	rows, err := h.store.Queries.ListTemplates(ctx, sqlc.ListTemplatesParams{Limit: limit, Offset: offset})
	if err != nil {
		return api.ListTemplates500JSONResponse{InternalErrorJSONResponse: api.InternalErrorJSONResponse{Message: err.Error()}}, nil
	}

	data := make([]api.Template, len(rows))
	for i, r := range rows {
		data[i] = templateToAPI(r)
	}
	return api.ListTemplates200JSONResponse{Data: data, Total: len(data)}, nil
}

func (h *Handler) CreateTemplate(ctx context.Context, req api.CreateTemplateRequestObject) (api.CreateTemplateResponseObject, error) {
	schema := mapToJSON(nil)
	if req.Body.Schema != nil {
		schema = mapToJSON(*req.Body.Schema)
	}

	row, err := h.store.Queries.CreateTemplate(ctx, sqlc.CreateTemplateParams{
		Slug:        req.Body.Slug,
		Name:        req.Body.Name,
		Description: req.Body.Description,
		Schema:      schema,
	})
	if err != nil {
		return api.CreateTemplate500JSONResponse{InternalErrorJSONResponse: api.InternalErrorJSONResponse{Message: err.Error()}}, nil
	}
	return api.CreateTemplate201JSONResponse(templateToAPI(row)), nil
}

func (h *Handler) GetTemplate(ctx context.Context, req api.GetTemplateRequestObject) (api.GetTemplateResponseObject, error) {
	row, err := h.store.Queries.GetTemplate(ctx, pgtype.UUID{Bytes: req.Id, Valid: true})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return api.GetTemplate404JSONResponse{NotFoundJSONResponse: api.NotFoundJSONResponse{Message: "template not found"}}, nil
		}
		return api.GetTemplate500JSONResponse{InternalErrorJSONResponse: api.InternalErrorJSONResponse{Message: err.Error()}}, nil
	}
	return api.GetTemplate200JSONResponse(templateToAPI(row)), nil
}

func (h *Handler) UpdateTemplate(ctx context.Context, req api.UpdateTemplateRequestObject) (api.UpdateTemplateResponseObject, error) {
	schema := mapToJSON(nil)
	if req.Body.Schema != nil {
		schema = mapToJSON(*req.Body.Schema)
	}

	row, err := h.store.Queries.UpdateTemplate(ctx, sqlc.UpdateTemplateParams{
		ID:          pgtype.UUID{Bytes: req.Id, Valid: true},
		Slug:        req.Body.Slug,
		Name:        req.Body.Name,
		Description: req.Body.Description,
		Schema:      schema,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return api.UpdateTemplate404JSONResponse{NotFoundJSONResponse: api.NotFoundJSONResponse{Message: "template not found"}}, nil
		}
		return api.UpdateTemplate500JSONResponse{InternalErrorJSONResponse: api.InternalErrorJSONResponse{Message: err.Error()}}, nil
	}
	return api.UpdateTemplate200JSONResponse(templateToAPI(row)), nil
}

func (h *Handler) DeleteTemplate(ctx context.Context, req api.DeleteTemplateRequestObject) (api.DeleteTemplateResponseObject, error) {
	err := h.store.Queries.DeleteTemplate(ctx, pgtype.UUID{Bytes: req.Id, Valid: true})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return api.DeleteTemplate404JSONResponse{NotFoundJSONResponse: api.NotFoundJSONResponse{Message: "template not found"}}, nil
		}
		return api.DeleteTemplate500JSONResponse{InternalErrorJSONResponse: api.InternalErrorJSONResponse{Message: err.Error()}}, nil
	}
	return api.DeleteTemplate204Response{}, nil
}
