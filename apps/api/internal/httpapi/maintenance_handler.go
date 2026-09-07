package httpapi

import (
	"errors"
	"net/http"

	"github.com/RealEstateSassApplication/property-management-OS/apps/api/internal/maintenance"
)

type maintenanceHandler struct {
	service *maintenance.Service
}

func (h maintenanceHandler) listVendors(w http.ResponseWriter, r *http.Request) {
	organizationID, _ := organizationIDFromContext(r.Context())
	items, err := h.service.ListVendors(r.Context(), organizationID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "vendors_list_failed", "could not load vendors")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"data": items})
}

func (h maintenanceHandler) createVendor(w http.ResponseWriter, r *http.Request) {
	var input maintenance.CreateVendorInput
	if err := decodeJSON(r, &input); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_json", err.Error())
		return
	}
	organizationID, _ := organizationIDFromContext(r.Context())
	item, err := h.service.CreateVendor(r.Context(), organizationID, input)
	if err != nil {
		writeError(w, http.StatusBadRequest, "vendor_invalid", err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, map[string]any{"data": item})
}

func (h maintenanceHandler) listRequests(w http.ResponseWriter, r *http.Request) {
	organizationID, _ := organizationIDFromContext(r.Context())
	items, err := h.service.ListRequests(r.Context(), organizationID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "maintenance_requests_list_failed", "could not load maintenance requests")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"data": items})
}

func (h maintenanceHandler) createRequest(w http.ResponseWriter, r *http.Request) {
	var input maintenance.CreateRequestInput
	if err := decodeJSON(r, &input); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_json", err.Error())
		return
	}
	organizationID, _ := organizationIDFromContext(r.Context())
	userID, _ := userIDFromContext(r.Context())
	item, err := h.service.CreateRequest(r.Context(), organizationID, userID, input)
	switch {
	case errors.Is(err, maintenance.ErrPropertyNotFound), errors.Is(err, maintenance.ErrUnitNotFound), errors.Is(err, maintenance.ErrTenantNotFound):
		writeError(w, http.StatusNotFound, "maintenance_resource_not_found", err.Error())
		return
	case errors.Is(err, maintenance.ErrTenantNotOccupant):
		writeError(w, http.StatusConflict, "tenant_occupancy_mismatch", err.Error())
		return
	case err != nil:
		writeError(w, http.StatusBadRequest, "maintenance_request_invalid", err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, map[string]any{"data": item})
}

func (h maintenanceHandler) updateRequestStatus(w http.ResponseWriter, r *http.Request) {
	var input maintenance.UpdateRequestStatusInput
	if err := decodeJSON(r, &input); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_json", err.Error())
		return
	}
	organizationID, _ := organizationIDFromContext(r.Context())
	item, err := h.service.UpdateRequestStatus(r.Context(), organizationID, r.PathValue("requestID"), input)
	switch {
	case errors.Is(err, maintenance.ErrRequestNotFound):
		writeError(w, http.StatusNotFound, "maintenance_request_not_found", err.Error())
		return
	case errors.Is(err, maintenance.ErrInvalidRequestTransition):
		writeError(w, http.StatusConflict, "maintenance_request_transition_conflict", err.Error())
		return
	case err != nil:
		writeError(w, http.StatusBadRequest, "maintenance_request_invalid", err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"data": item})
}

func (h maintenanceHandler) listWorkOrders(w http.ResponseWriter, r *http.Request) {
	organizationID, _ := organizationIDFromContext(r.Context())
	items, err := h.service.ListWorkOrders(r.Context(), organizationID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "work_orders_list_failed", "could not load work orders")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"data": items})
}

func (h maintenanceHandler) createWorkOrder(w http.ResponseWriter, r *http.Request) {
	var input maintenance.CreateWorkOrderInput
	if err := decodeJSON(r, &input); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_json", err.Error())
		return
	}
	organizationID, _ := organizationIDFromContext(r.Context())
	item, err := h.service.CreateWorkOrder(r.Context(), organizationID, input)
	switch {
	case errors.Is(err, maintenance.ErrRequestNotFound), errors.Is(err, maintenance.ErrVendorNotFound):
		writeError(w, http.StatusNotFound, "work_order_resource_not_found", err.Error())
		return
	case errors.Is(err, maintenance.ErrRequestClosed), errors.Is(err, maintenance.ErrVendorInactive):
		writeError(w, http.StatusConflict, "work_order_conflict", err.Error())
		return
	case err != nil:
		writeError(w, http.StatusBadRequest, "work_order_invalid", err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, map[string]any{"data": item})
}

func (h maintenanceHandler) updateWorkOrderStatus(w http.ResponseWriter, r *http.Request) {
	var input maintenance.UpdateWorkOrderStatusInput
	if err := decodeJSON(r, &input); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_json", err.Error())
		return
	}
	organizationID, _ := organizationIDFromContext(r.Context())
	item, err := h.service.UpdateWorkOrderStatus(r.Context(), organizationID, r.PathValue("workOrderID"), input)
	switch {
	case errors.Is(err, maintenance.ErrWorkOrderNotFound):
		writeError(w, http.StatusNotFound, "work_order_not_found", err.Error())
		return
	case errors.Is(err, maintenance.ErrInvalidWorkOrderTransition), errors.Is(err, maintenance.ErrCompletionEvidenceRequired):
		writeError(w, http.StatusConflict, "work_order_transition_conflict", err.Error())
		return
	case err != nil:
		writeError(w, http.StatusBadRequest, "work_order_invalid", err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"data": item})
}

func (h maintenanceHandler) listQuotes(w http.ResponseWriter, r *http.Request) {
	organizationID, _ := organizationIDFromContext(r.Context())
	items, err := h.service.ListQuotes(r.Context(), organizationID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "maintenance_quotes_list_failed", "could not load maintenance quotes")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"data": items})
}

func (h maintenanceHandler) createQuote(w http.ResponseWriter, r *http.Request) {
	var input maintenance.CreateQuoteInput
	if err := decodeJSON(r, &input); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_json", err.Error())
		return
	}
	organizationID, _ := organizationIDFromContext(r.Context())
	item, err := h.service.CreateQuote(r.Context(), organizationID, input)
	switch {
	case errors.Is(err, maintenance.ErrWorkOrderNotFound), errors.Is(err, maintenance.ErrVendorNotFound):
		writeError(w, http.StatusNotFound, "quote_resource_not_found", err.Error())
		return
	case errors.Is(err, maintenance.ErrWorkOrderClosed), errors.Is(err, maintenance.ErrVendorInactive):
		writeError(w, http.StatusConflict, "quote_conflict", err.Error())
		return
	case err != nil:
		writeError(w, http.StatusBadRequest, "quote_invalid", err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, map[string]any{"data": item})
}

func (h maintenanceHandler) decideQuote(w http.ResponseWriter, r *http.Request) {
	var input maintenance.DecideQuoteInput
	if err := decodeJSON(r, &input); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_json", err.Error())
		return
	}
	organizationID, _ := organizationIDFromContext(r.Context())
	userID, _ := userIDFromContext(r.Context())
	item, err := h.service.DecideQuote(r.Context(), organizationID, r.PathValue("quoteID"), userID, input)
	switch {
	case errors.Is(err, maintenance.ErrQuoteNotFound):
		writeError(w, http.StatusNotFound, "maintenance_quote_not_found", err.Error())
		return
	case errors.Is(err, maintenance.ErrQuoteNotSubmitted), errors.Is(err, maintenance.ErrApprovedQuoteExists), errors.Is(err, maintenance.ErrWorkOrderClosed):
		writeError(w, http.StatusConflict, "quote_decision_conflict", err.Error())
		return
	case err != nil:
		writeError(w, http.StatusBadRequest, "quote_decision_invalid", err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"data": item})
}

func (h maintenanceHandler) listEvidence(w http.ResponseWriter, r *http.Request) {
	organizationID, _ := organizationIDFromContext(r.Context())
	items, err := h.service.ListEvidence(r.Context(), organizationID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "maintenance_evidence_list_failed", "could not load completion evidence")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"data": items})
}

func (h maintenanceHandler) createEvidence(w http.ResponseWriter, r *http.Request) {
	var input maintenance.CreateEvidenceInput
	if err := decodeJSON(r, &input); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_json", err.Error())
		return
	}
	organizationID, _ := organizationIDFromContext(r.Context())
	userID, _ := userIDFromContext(r.Context())
	item, err := h.service.CreateEvidence(r.Context(), organizationID, userID, input)
	switch {
	case errors.Is(err, maintenance.ErrWorkOrderNotFound):
		writeError(w, http.StatusNotFound, "work_order_not_found", err.Error())
		return
	case errors.Is(err, maintenance.ErrWorkOrderClosed):
		writeError(w, http.StatusConflict, "work_order_closed", err.Error())
		return
	case err != nil:
		writeError(w, http.StatusBadRequest, "maintenance_evidence_invalid", err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, map[string]any{"data": item})
}
