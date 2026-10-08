package httpapi

import (
	"bytes"
	"errors"
	"io"
	"mime"
	"net/http"

	"github.com/RealEstateSassApplication/property-management-OS/apps/api/internal/onboarding"
)

const maxPortfolioCSVBytes = 512 * 1024

// previewPropertyCSV is deliberately read-only. It never creates/updates
// property records, changes accounting balances, or stores an uploaded CSV.
func previewPropertyCSV(w http.ResponseWriter, r *http.Request) {
	mediaType, _, err := mime.ParseMediaType(r.Header.Get("Content-Type"))
	if err != nil || mediaType != "text/csv" {
		writeError(w, http.StatusUnsupportedMediaType, "csv_required", "send a text/csv file")
		return
	}
	data, err := io.ReadAll(io.LimitReader(r.Body, maxPortfolioCSVBytes+1))
	if err != nil {
		writeError(w, http.StatusBadRequest, "csv_read_failed", "could not read CSV")
		return
	}
	if len(data) > maxPortfolioCSVBytes {
		writeError(w, http.StatusRequestEntityTooLarge, "csv_too_large", "CSV exceeds 512 KiB")
		return
	}
	report, err := onboarding.PreviewPropertiesCSV(bytes.NewReader(data))
	switch {
	case errors.Is(err, onboarding.ErrTooManyProperties):
		writeError(w, http.StatusRequestEntityTooLarge, "too_many_properties", "CSV contains more than 500 properties")
	case errors.Is(err, onboarding.ErrInvalidHeader), errors.Is(err, onboarding.ErrInvalidCSV), errors.Is(err, onboarding.ErrNoProperties):
		writeError(w, http.StatusBadRequest, "invalid_csv", err.Error())
	case err != nil:
		writeError(w, http.StatusBadRequest, "invalid_csv", "could not parse property CSV")
	default:
		writeJSON(w, http.StatusOK, map[string]any{"data": report})
	}
}
