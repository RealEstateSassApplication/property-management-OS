package httpapi

import (
	"errors"
	"net/http"

	"github.com/RealEstateSassApplication/property-management-OS/apps/api/internal/notifications"
)

type notificationHandler struct{ service *notifications.Service }

func (h notificationHandler) list(w http.ResponseWriter, r *http.Request) {
	organizationID, _ := organizationIDFromContext(r.Context())
	items, err := h.service.List(r.Context(), organizationID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "notifications_list_failed", "could not load notifications")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"data": items})
}

func (h notificationHandler) enqueue(w http.ResponseWriter, r *http.Request) {
	var input notifications.EnqueueInput
	if err := decodeJSON(r, &input); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_json", err.Error())
		return
	}
	organizationID, _ := organizationIDFromContext(r.Context())
	userID, _ := userIDFromContext(r.Context())
	item, err := h.service.Enqueue(r.Context(), organizationID, userID, input)
	if err != nil {
		writeError(w, http.StatusBadRequest, "notification_invalid", err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, map[string]any{"data": item})
}

func (h notificationHandler) queueRentReminder(w http.ResponseWriter, r *http.Request) {
	var input notifications.RentReminderInput
	if err := decodeJSON(r, &input); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_json", err.Error())
		return
	}
	organizationID, _ := organizationIDFromContext(r.Context())
	userID, _ := userIDFromContext(r.Context())
	item, err := h.service.QueueRentReminder(r.Context(), organizationID, userID, input)
	switch {
	case errors.Is(err, notifications.ErrObligationNotFound):
		writeError(w, http.StatusNotFound, "rent_obligation_not_found", err.Error())
		return
	case errors.Is(err, notifications.ErrRentReminderNotApplicable):
		writeError(w, http.StatusConflict, "rent_reminder_not_applicable", err.Error())
		return
	case err != nil:
		writeError(w, http.StatusBadRequest, "rent_reminder_invalid", err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, map[string]any{"data": item})
}
