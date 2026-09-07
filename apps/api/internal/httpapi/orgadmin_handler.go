package httpapi

import (
	"errors"
	"net/http"

	"github.com/RealEstateSassApplication/property-management-OS/apps/api/internal/orgadmin"
)

type orgAdminHandler struct{ service *orgadmin.Service }

func (h orgAdminHandler) getSettings(w http.ResponseWriter, r *http.Request) {
	organizationID, _ := organizationIDFromContext(r.Context())
	item, err := h.service.GetSettings(r.Context(), organizationID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "organization_load_failed", "could not load organization settings")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"data": item})
}

func (h orgAdminHandler) updateSettings(w http.ResponseWriter, r *http.Request) {
	var input orgadmin.UpdateOrganizationInput
	if err := decodeJSON(r, &input); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_json", err.Error())
		return
	}
	organizationID, _ := organizationIDFromContext(r.Context())
	item, err := h.service.UpdateSettings(r.Context(), organizationID, input)
	if err != nil {
		writeError(w, http.StatusBadRequest, "organization_settings_invalid", err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"data": item})
}

func (h orgAdminHandler) listMembers(w http.ResponseWriter, r *http.Request) {
	organizationID, _ := organizationIDFromContext(r.Context())
	items, err := h.service.ListMembers(r.Context(), organizationID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "members_list_failed", "could not load organization members")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"data": items})
}

func (h orgAdminHandler) createMember(w http.ResponseWriter, r *http.Request) {
	var input orgadmin.CreateMemberInput
	if err := decodeJSON(r, &input); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_json", err.Error())
		return
	}
	organizationID, _ := organizationIDFromContext(r.Context())
	item, err := h.service.CreateMember(r.Context(), organizationID, input)
	if err != nil {
		writeError(w, http.StatusBadRequest, "member_invalid", err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, map[string]any{"data": item})
}

func (h orgAdminHandler) updateMemberRole(w http.ResponseWriter, r *http.Request) {
	var input orgadmin.UpdateMemberRoleInput
	if err := decodeJSON(r, &input); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_json", err.Error())
		return
	}
	organizationID, _ := organizationIDFromContext(r.Context())
	item, err := h.service.UpdateMemberRole(r.Context(), organizationID, r.PathValue("userID"), input)
	switch {
	case errors.Is(err, orgadmin.ErrMemberNotFound):
		writeError(w, http.StatusNotFound, "member_not_found", err.Error())
		return
	case errors.Is(err, orgadmin.ErrLastAdmin):
		writeError(w, http.StatusConflict, "last_admin", err.Error())
		return
	case err != nil:
		writeError(w, http.StatusBadRequest, "member_role_invalid", err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"data": item})
}

func (h orgAdminHandler) removeMember(w http.ResponseWriter, r *http.Request) {
	organizationID, _ := organizationIDFromContext(r.Context())
	err := h.service.RemoveMember(r.Context(), organizationID, r.PathValue("userID"))
	switch {
	case errors.Is(err, orgadmin.ErrMemberNotFound):
		writeError(w, http.StatusNotFound, "member_not_found", err.Error())
		return
	case errors.Is(err, orgadmin.ErrLastAdmin):
		writeError(w, http.StatusConflict, "last_admin", err.Error())
		return
	case err != nil:
		writeError(w, http.StatusInternalServerError, "member_remove_failed", "could not remove organization member")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h orgAdminHandler) listPortalLinks(w http.ResponseWriter, r *http.Request) {
	organizationID, _ := organizationIDFromContext(r.Context())
	items, err := h.service.ListPortalLinks(r.Context(), organizationID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "portal_links_failed", "could not load portal links")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"data": items})
}

func (h orgAdminHandler) linkOwner(w http.ResponseWriter, r *http.Request) {
	var input orgadmin.CreateOwnerPortalLinkInput
	if err := decodeJSON(r, &input); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_json", err.Error())
		return
	}
	organizationID, _ := organizationIDFromContext(r.Context())
	item, err := h.service.LinkOwner(r.Context(), organizationID, input)
	if err != nil {
		writeError(w, http.StatusBadRequest, "owner_portal_link_invalid", err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, map[string]any{"data": item})
}

func (h orgAdminHandler) linkTenant(w http.ResponseWriter, r *http.Request) {
	var input orgadmin.CreateTenantPortalLinkInput
	if err := decodeJSON(r, &input); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_json", err.Error())
		return
	}
	organizationID, _ := organizationIDFromContext(r.Context())
	item, err := h.service.LinkTenant(r.Context(), organizationID, input)
	if err != nil {
		writeError(w, http.StatusBadRequest, "tenant_portal_link_invalid", err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, map[string]any{"data": item})
}

func (h orgAdminHandler) unlinkOwner(w http.ResponseWriter, r *http.Request) {
	organizationID, _ := organizationIDFromContext(r.Context())
	if err := h.service.UnlinkOwner(r.Context(), organizationID, r.PathValue("userID"), r.PathValue("ownerID")); err != nil {
		writeError(w, http.StatusNotFound, "owner_portal_link_not_found", err.Error())
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h orgAdminHandler) unlinkTenant(w http.ResponseWriter, r *http.Request) {
	organizationID, _ := organizationIDFromContext(r.Context())
	if err := h.service.UnlinkTenant(r.Context(), organizationID, r.PathValue("userID"), r.PathValue("tenantID")); err != nil {
		writeError(w, http.StatusNotFound, "tenant_portal_link_not_found", err.Error())
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
