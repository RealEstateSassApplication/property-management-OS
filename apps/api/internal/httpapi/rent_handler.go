package httpapi

import (
	"errors"
	"net/http"

	"github.com/RealEstateSassApplication/property-management-OS/apps/api/internal/rent"
)

type rentHandler struct {
	service *rent.Service
}

func (h rentHandler) listObligations(w http.ResponseWriter, r *http.Request) {
	organizationID, _ := organizationIDFromContext(r.Context())
	items, err := h.service.ListObligations(r.Context(), organizationID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "rent_obligations_list_failed", "could not load rent obligations")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"data": items})
}

func (h rentHandler) createObligation(w http.ResponseWriter, r *http.Request) {
	var input rent.CreateObligationInput
	if err := decodeJSON(r, &input); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_json", err.Error())
		return
	}
	organizationID, _ := organizationIDFromContext(r.Context())
	item, err := h.service.CreateObligation(r.Context(), organizationID, input)
	switch {
	case errors.Is(err, rent.ErrLeaseNotFound):
		writeError(w, http.StatusNotFound, "lease_not_found", err.Error())
		return
	case errors.Is(err, rent.ErrObligationExists):
		writeError(w, http.StatusConflict, "rent_obligation_exists", err.Error())
		return
	case errors.Is(err, rent.ErrLeaseNotActive), errors.Is(err, rent.ErrPeriodOutsideLease):
		writeError(w, http.StatusBadRequest, "rent_obligation_invalid", err.Error())
		return
	case err != nil:
		writeError(w, http.StatusBadRequest, "rent_obligation_invalid", err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, map[string]any{"data": item})
}

func (h rentHandler) listPayments(w http.ResponseWriter, r *http.Request) {
	organizationID, _ := organizationIDFromContext(r.Context())
	items, err := h.service.ListPayments(r.Context(), organizationID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "payments_list_failed", "could not load payments")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"data": items})
}

func (h rentHandler) createPayment(w http.ResponseWriter, r *http.Request) {
	var input rent.CreatePaymentInput
	if err := decodeJSON(r, &input); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_json", err.Error())
		return
	}
	organizationID, _ := organizationIDFromContext(r.Context())
	item, err := h.service.CreatePayment(r.Context(), organizationID, input)
	if errors.Is(err, rent.ErrTenantNotFound) {
		writeError(w, http.StatusNotFound, "tenant_not_found", err.Error())
		return
	}
	if err != nil {
		writeError(w, http.StatusBadRequest, "payment_invalid", err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, map[string]any{"data": item})
}

func (h rentHandler) createAllocation(w http.ResponseWriter, r *http.Request) {
	var input rent.CreateAllocationInput
	if err := decodeJSON(r, &input); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_json", err.Error())
		return
	}
	organizationID, _ := organizationIDFromContext(r.Context())
	item, err := h.service.CreateAllocation(r.Context(), organizationID, input)
	switch {
	case errors.Is(err, rent.ErrPaymentNotFound), errors.Is(err, rent.ErrObligationNotFound):
		writeError(w, http.StatusNotFound, "allocation_resource_not_found", err.Error())
		return
	case errors.Is(err, rent.ErrAllocationExceedsPayment), errors.Is(err, rent.ErrAllocationExceedsObligation):
		writeError(w, http.StatusConflict, "allocation_balance_conflict", err.Error())
		return
	case errors.Is(err, rent.ErrCurrencyMismatch), errors.Is(err, rent.ErrTenantMismatch), errors.Is(err, rent.ErrPaymentNotPosted), errors.Is(err, rent.ErrObligationVoided):
		writeError(w, http.StatusBadRequest, "allocation_invalid", err.Error())
		return
	case err != nil:
		writeError(w, http.StatusBadRequest, "allocation_invalid", err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, map[string]any{"data": item})
}
