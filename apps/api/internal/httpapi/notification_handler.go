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

func (h notificationHandler) registerPushDevice(w http.ResponseWriter, r *http.Request) {
	var input notifications.RegisterPushDeviceInput
	if err := decodeJSON(r, &input); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_json", err.Error())
		return
	}
	organizationID, _ := organizationIDFromContext(r.Context())
	userID, _ := userIDFromContext(r.Context())
	item, err := h.service.RegisterPushDevice(r.Context(), organizationID, userID, input)
	if err != nil {
		writeError(w, http.StatusBadRequest, "push_device_invalid", err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, map[string]any{"data": item})
}

func (h notificationHandler) listPushDevices(w http.ResponseWriter, r *http.Request) {
	organizationID, _ := organizationIDFromContext(r.Context())
	userID, _ := userIDFromContext(r.Context())
	items, err := h.service.ListPushDevices(r.Context(), organizationID, userID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "push_devices_list_failed", "could not load push devices")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"data": items})
}

func (h notificationHandler) deletePushDevice(w http.ResponseWriter, r *http.Request) {
	organizationID, _ := organizationIDFromContext(r.Context())
	userID, _ := userIDFromContext(r.Context())
	err := h.service.DeletePushDevice(r.Context(), organizationID, userID, r.PathValue("deviceID"))
	if errors.Is(err, notifications.ErrPushDeviceNotFound) {
		writeError(w, http.StatusNotFound, "push_device_not_found", err.Error())
		return
	}
	if err != nil {
		writeError(w, http.StatusBadRequest, "push_device_delete_failed", err.Error())
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h notificationHandler) queuePushToUser(w http.ResponseWriter, r *http.Request) {
	var input notifications.PushToUserInput
	if err := decodeJSON(r, &input); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_json", err.Error())
		return
	}
	organizationID, _ := organizationIDFromContext(r.Context())
	userID, _ := userIDFromContext(r.Context())
	items, err := h.service.QueuePushToUser(r.Context(), organizationID, userID, input)
	if errors.Is(err, notifications.ErrPushRecipientUnavailable) {
		writeError(w, http.StatusConflict, "push_recipient_unavailable", err.Error())
		return
	}
	if err != nil {
		writeError(w, http.StatusBadRequest, "push_notification_invalid", err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, map[string]any{"data": items})
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
