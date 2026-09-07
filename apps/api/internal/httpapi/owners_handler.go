package httpapi

import (
	"errors"
	"net/http"

	"github.com/RealEstateSassApplication/property-management-OS/apps/api/internal/owners"
)

type ownerHandler struct {
	service *owners.Service
}

func (h ownerHandler) list(w http.ResponseWriter, r *http.Request) {
	organizationID, _ := organizationIDFromContext(r.Context())
	items, err := h.service.List(r.Context(), organizationID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "owners_list_failed", "could not load owners")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"data": items})
}

func (h ownerHandler) get(w http.ResponseWriter, r *http.Request) {
	organizationID, _ := organizationIDFromContext(r.Context())
	item, err := h.service.Get(r.Context(), organizationID, r.PathValue("ownerID"))
	if errors.Is(err, owners.ErrNotFound) {
		writeError(w, http.StatusNotFound, "owner_not_found", "owner not found")
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, "owner_get_failed", "could not load owner")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"data": item})
}

func (h ownerHandler) create(w http.ResponseWriter, r *http.Request) {
	var input owners.CreateOwnerInput
	if err := decodeJSON(r, &input); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_json", err.Error())
		return
	}
	organizationID, _ := organizationIDFromContext(r.Context())
	item, err := h.service.Create(r.Context(), organizationID, input)
	if err != nil {
		writeError(w, http.StatusBadRequest, "owner_invalid", err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, map[string]any{"data": item})
}

func (h ownerHandler) listInterests(w http.ResponseWriter, r *http.Request) {
	organizationID, _ := organizationIDFromContext(r.Context())
	items, err := h.service.ListInterests(r.Context(), organizationID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "ownership_list_failed", "could not load ownership interests")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"data": items})
}

func (h ownerHandler) createInterest(w http.ResponseWriter, r *http.Request) {
	var input owners.CreateInterestInput
	if err := decodeJSON(r, &input); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_json", err.Error())
		return
	}
	organizationID, _ := organizationIDFromContext(r.Context())
	item, err := h.service.CreateInterest(r.Context(), organizationID, input)
	switch {
	case errors.Is(err, owners.ErrOwnerNotFound), errors.Is(err, owners.ErrPropertyNotFound):
		writeError(w, http.StatusNotFound, "ownership_resource_not_found", err.Error())
		return
	case errors.Is(err, owners.ErrOwnershipExceeded), errors.Is(err, owners.ErrCurrentInterestExists):
		writeError(w, http.StatusConflict, "ownership_conflict", err.Error())
		return
	case err != nil:
		writeError(w, http.StatusBadRequest, "ownership_invalid", err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, map[string]any{"data": item})
}
