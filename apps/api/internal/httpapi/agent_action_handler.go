package httpapi

import (
	"errors"
	"net/http"

	"github.com/RealEstateSassApplication/property-management-OS/apps/api/internal/agentactions"
)

type agentActionHandler struct{ service *agentactions.Service }

func (h agentActionHandler) list(w http.ResponseWriter, r *http.Request) {
	organizationID, _ := organizationIDFromContext(r.Context())
	items, err := h.service.List(r.Context(), organizationID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "agent_actions_list_failed", "could not load agent action requests")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"data": items})
}

func (h agentActionHandler) proposeQuoteApproval(w http.ResponseWriter, r *http.Request) {
	var input agentactions.ProposeQuoteApprovalInput
	if err := decodeJSON(r, &input); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_json", err.Error())
		return
	}
	organizationID, _ := organizationIDFromContext(r.Context())
	userID, _ := userIDFromContext(r.Context())
	item, err := h.service.ProposeMaintenanceQuoteApproval(r.Context(), organizationID, userID, input)
	switch {
	case errors.Is(err, agentactions.ErrQuoteNotFound):
		writeError(w, http.StatusNotFound, "maintenance_quote_not_found", err.Error())
		return
	case errors.Is(err, agentactions.ErrQuoteNotProposable):
		writeError(w, http.StatusConflict, "maintenance_quote_not_proposable", err.Error())
		return
	case err != nil:
		writeError(w, http.StatusBadRequest, "agent_action_invalid", err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, map[string]any{"data": item})
}

func (h agentActionHandler) decide(w http.ResponseWriter, r *http.Request) {
	var input agentactions.DecisionInput
	if err := decodeJSON(r, &input); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_json", err.Error())
		return
	}
	organizationID, _ := organizationIDFromContext(r.Context())
	userID, _ := userIDFromContext(r.Context())
	item, err := h.service.Decide(r.Context(), organizationID, r.PathValue("actionID"), userID, input)
	switch {
	case errors.Is(err, agentactions.ErrActionNotFound):
		writeError(w, http.StatusNotFound, "agent_action_not_found", err.Error())
		return
	case errors.Is(err, agentactions.ErrActionNotPending):
		writeError(w, http.StatusConflict, "agent_action_not_pending", err.Error())
		return
	case errors.Is(err, agentactions.ErrExecutionFailed):
		writeJSON(w, http.StatusConflict, map[string]any{"data": item, "error": map[string]string{"code": "agent_action_execution_failed", "message": err.Error()}})
		return
	case err != nil:
		writeError(w, http.StatusBadRequest, "agent_action_decision_invalid", err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"data": item})
}
