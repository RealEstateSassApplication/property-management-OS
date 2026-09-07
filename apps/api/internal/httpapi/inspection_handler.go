package httpapi

import (
	"errors"
	"net/http"

	"github.com/RealEstateSassApplication/property-management-OS/apps/api/internal/inspections"
)

type inspectionHandler struct{ service *inspections.Service }

func (h inspectionHandler) list(w http.ResponseWriter, r *http.Request) {
	organizationID, _ := organizationIDFromContext(r.Context())
	items, err := h.service.List(r.Context(), organizationID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "inspections_list_failed", "could not load inspections")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"data": items})
}

func (h inspectionHandler) get(w http.ResponseWriter, r *http.Request) {
	organizationID, _ := organizationIDFromContext(r.Context())
	item, err := h.service.Get(r.Context(), organizationID, r.PathValue("inspectionID"))
	if errors.Is(err, inspections.ErrInspectionNotFound) {
		writeError(w, http.StatusNotFound, "inspection_not_found", err.Error())
		return
	}
	if err != nil {
		writeError(w, http.StatusBadRequest, "inspection_get_failed", err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"data": item})
}

func (h inspectionHandler) listItems(w http.ResponseWriter, r *http.Request) {
	organizationID, _ := organizationIDFromContext(r.Context())
	items, err := h.service.ListItems(r.Context(), organizationID, r.PathValue("inspectionID"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "inspection_items_list_failed", err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"data": items})
}

func (h inspectionHandler) create(w http.ResponseWriter, r *http.Request) {
	var input inspections.CreateInspectionInput
	if err := decodeJSON(r, &input); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_json", err.Error())
		return
	}
	organizationID, _ := organizationIDFromContext(r.Context())
	actorUserID, _ := userIDFromContext(r.Context())
	item, err := h.service.Create(r.Context(), organizationID, actorUserID, input)
	switch {
	case errors.Is(err, inspections.ErrTenancyNotFound):
		writeError(w, http.StatusNotFound, "tenancy_not_found", err.Error())
		return
	case errors.Is(err, inspections.ErrInvalidInspectionType):
		writeError(w, http.StatusBadRequest, "inspection_type_invalid", err.Error())
		return
	case err != nil:
		writeError(w, http.StatusBadRequest, "inspection_invalid", err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, map[string]any{"data": item})
}

func (h inspectionHandler) createItem(w http.ResponseWriter, r *http.Request) {
	var input inspections.CreateItemInput
	if err := decodeJSON(r, &input); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_json", err.Error())
		return
	}
	input.InspectionID = r.PathValue("inspectionID")
	organizationID, _ := organizationIDFromContext(r.Context())
	item, err := h.service.CreateItem(r.Context(), organizationID, input)
	switch {
	case errors.Is(err, inspections.ErrInspectionNotFound), errors.Is(err, inspections.ErrDocumentNotFound):
		writeError(w, http.StatusNotFound, "inspection_resource_not_found", err.Error())
		return
	case errors.Is(err, inspections.ErrInspectionClosed), errors.Is(err, inspections.ErrDocumentNotAvailable):
		writeError(w, http.StatusConflict, "inspection_item_conflict", err.Error())
		return
	case err != nil:
		writeError(w, http.StatusBadRequest, "inspection_item_invalid", err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, map[string]any{"data": item})
}

func (h inspectionHandler) complete(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Summary string `json:"summary"`
	}
	if err := decodeJSON(r, &body); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_json", err.Error())
		return
	}
	organizationID, _ := organizationIDFromContext(r.Context())
	actorUserID, _ := userIDFromContext(r.Context())
	item, err := h.service.Complete(r.Context(), organizationID, actorUserID, inspections.CompleteInspectionInput{InspectionID: r.PathValue("inspectionID"), Summary: body.Summary})
	switch {
	case errors.Is(err, inspections.ErrInspectionNotFound):
		writeError(w, http.StatusNotFound, "inspection_not_found", err.Error())
		return
	case errors.Is(err, inspections.ErrInspectionClosed), errors.Is(err, inspections.ErrInspectionHasNoItems):
		writeError(w, http.StatusConflict, "inspection_complete_conflict", err.Error())
		return
	case err != nil:
		writeError(w, http.StatusBadRequest, "inspection_complete_invalid", err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"data": item})
}

func (h inspectionHandler) acknowledge(w http.ResponseWriter, r *http.Request) {
	organizationID, _ := organizationIDFromContext(r.Context())
	actorUserID, _ := userIDFromContext(r.Context())
	item, err := h.service.Acknowledge(r.Context(), organizationID, actorUserID, inspections.AcknowledgeInspectionInput{InspectionID: r.PathValue("inspectionID")})
	switch {
	case errors.Is(err, inspections.ErrInspectionNotFound):
		writeError(w, http.StatusNotFound, "inspection_not_found", err.Error())
		return
	case errors.Is(err, inspections.ErrInspectionNotCompleted):
		writeError(w, http.StatusConflict, "inspection_not_completed", err.Error())
		return
	case err != nil:
		writeError(w, http.StatusBadRequest, "inspection_acknowledge_invalid", err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"data": item})
}
