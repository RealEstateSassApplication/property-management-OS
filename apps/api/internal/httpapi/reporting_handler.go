package httpapi

import (
	"net/http"
	"strconv"

	"github.com/RealEstateSassApplication/property-management-OS/apps/api/internal/reporting"
)

type reportingHandler struct{ service *reporting.Service }

func (h reportingHandler) dashboard(w http.ResponseWriter, r *http.Request) {
	organizationID, _ := organizationIDFromContext(r.Context())
	item, err := h.service.Dashboard(r.Context(), organizationID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "reporting_failed", "could not load reporting dashboard")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"data": item})
}

func (h reportingHandler) audit(w http.ResponseWriter, r *http.Request) {
	limit := 100
	if raw := r.URL.Query().Get("limit"); raw != "" {
		parsed, err := strconv.Atoi(raw)
		if err != nil {
			writeError(w, http.StatusBadRequest, "invalid_limit", "limit must be an integer")
			return
		}
		limit = parsed
	}
	organizationID, _ := organizationIDFromContext(r.Context())
	items, err := h.service.ListAudit(r.Context(), organizationID, limit)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "audit_failed", "could not load audit events")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"data": items})
}
