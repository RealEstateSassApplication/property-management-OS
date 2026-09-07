package httpapi

import (
	"errors"
	"net/http"

	"github.com/RealEstateSassApplication/property-management-OS/apps/api/internal/documents"
)

type documentHandler struct{ service *documents.Service }

func (h documentHandler) list(w http.ResponseWriter, r *http.Request) {
	organizationID, _ := organizationIDFromContext(r.Context())
	items, err := h.service.List(r.Context(), organizationID, documents.Filter{ResourceType: r.URL.Query().Get("resourceType"), ResourceID: r.URL.Query().Get("resourceId")})
	if err != nil {
		writeError(w, http.StatusInternalServerError, "documents_list_failed", "could not load documents")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"data": items})
}

func (h documentHandler) initiateUpload(w http.ResponseWriter, r *http.Request) {
	var input documents.InitiateUploadInput
	if err := decodeJSON(r, &input); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_json", err.Error())
		return
	}
	organizationID, _ := organizationIDFromContext(r.Context())
	userID, _ := userIDFromContext(r.Context())
	intent, err := h.service.InitiateUpload(r.Context(), organizationID, userID, input)
	switch {
	case errors.Is(err, documents.ErrResourceNotFound):
		writeError(w, http.StatusNotFound, "document_resource_not_found", err.Error())
	case errors.Is(err, documents.ErrStorageUnavailable):
		writeError(w, http.StatusServiceUnavailable, "document_storage_unavailable", err.Error())
	case err != nil:
		writeError(w, http.StatusBadRequest, "document_upload_invalid", err.Error())
	default:
		writeJSON(w, http.StatusCreated, map[string]any{"data": intent})
	}
}

func (h documentHandler) completeUpload(w http.ResponseWriter, r *http.Request) {
	organizationID, _ := organizationIDFromContext(r.Context())
	userID, _ := userIDFromContext(r.Context())
	item, err := h.service.CompleteUpload(r.Context(), organizationID, userID, r.PathValue("documentID"))
	switch {
	case errors.Is(err, documents.ErrNotFound):
		writeError(w, http.StatusNotFound, "document_not_found", err.Error())
	case errors.Is(err, documents.ErrUploadNotPending), errors.Is(err, documents.ErrObjectMismatch):
		writeError(w, http.StatusConflict, "document_upload_conflict", err.Error())
	case errors.Is(err, documents.ErrStorageUnavailable):
		writeError(w, http.StatusServiceUnavailable, "document_storage_unavailable", err.Error())
	case err != nil:
		writeError(w, http.StatusBadGateway, "document_storage_verification_failed", "could not verify uploaded object")
	default:
		writeJSON(w, http.StatusOK, map[string]any{"data": item})
	}
}

func (h documentHandler) download(w http.ResponseWriter, r *http.Request) {
	organizationID, _ := organizationIDFromContext(r.Context())
	grant, err := h.service.Download(r.Context(), organizationID, r.PathValue("documentID"))
	switch {
	case errors.Is(err, documents.ErrNotFound):
		writeError(w, http.StatusNotFound, "document_not_found", err.Error())
	case errors.Is(err, documents.ErrDocumentNotAvailable):
		writeError(w, http.StatusConflict, "document_not_available", err.Error())
	case errors.Is(err, documents.ErrStorageUnavailable):
		writeError(w, http.StatusServiceUnavailable, "document_storage_unavailable", err.Error())
	case err != nil:
		writeError(w, http.StatusBadGateway, "document_download_failed", "could not create secure download")
	default:
		writeJSON(w, http.StatusOK, map[string]any{"data": grant})
	}
}

func (h documentHandler) delete(w http.ResponseWriter, r *http.Request) {
	organizationID, _ := organizationIDFromContext(r.Context())
	userID, _ := userIDFromContext(r.Context())
	item, err := h.service.Delete(r.Context(), organizationID, userID, r.PathValue("documentID"))
	switch {
	case errors.Is(err, documents.ErrNotFound):
		writeError(w, http.StatusNotFound, "document_not_found", err.Error())
	case err != nil:
		writeError(w, http.StatusBadGateway, "document_delete_failed", "could not delete document")
	default:
		writeJSON(w, http.StatusOK, map[string]any{"data": item})
	}
}
