package httpapi

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/RealEstateSassApplication/property-management-OS/apps/api/internal/auth"
	"github.com/RealEstateSassApplication/property-management-OS/apps/api/internal/onboarding"
)

const testPortfolioCSV = "referenceCode,name,propertyType,addressLine1,city,countryCode\nP-01,Home,house,Main Street,Colombo,LK\n"

func previewRequest(body string) *http.Request {
	request := httptest.NewRequest(http.MethodPost, "/api/v1/onboarding/properties/preview", strings.NewReader(body))
	request.Header.Set("Content-Type", "text/csv")
	request.Header.Set("X-Organization-ID", "org-1")
	request.Header.Set("X-User-ID", "user-1")
	return request
}

func TestPortfolioPreviewProtectedAndReadOnly(t *testing.T) {
	authorized := auth.NewService(fakeAuthRepository{membership: auth.Membership{Role: "manager"}})
	handler := NewRouter(Dependencies{Authorization: authorized, AllowDevelopmentIdentity: true})
	result := httptest.NewRecorder()
	handler.ServeHTTP(result, previewRequest(testPortfolioCSV))
	if result.Code != http.StatusOK { t.Fatalf("expected 200, got %d: %s", result.Code, result.Body.String()) }
	var payload struct { Data onboarding.Report `json:"data"` }
	if err := json.Unmarshal(result.Body.Bytes(), &payload); err != nil { t.Fatal(err) }
	if payload.Data.ReadyRows != 1 || !payload.Data.CanImport { t.Fatalf("unexpected preview: %+v", payload.Data) }

	viewer := auth.NewService(fakeAuthRepository{membership: auth.Membership{Role: "viewer"}})
	viewerResult := httptest.NewRecorder()
	NewRouter(Dependencies{Authorization: viewer, AllowDevelopmentIdentity: true}).ServeHTTP(viewerResult, previewRequest(testPortfolioCSV))
	if viewerResult.Code != http.StatusForbidden { t.Fatalf("viewer should not preview import, got %d", viewerResult.Code) }

	missingOrganization := previewRequest(testPortfolioCSV)
	missingOrganization.Header.Del("X-Organization-ID")
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, missingOrganization)
	if response.Code != http.StatusBadRequest { t.Fatalf("missing organization header should fail, got %d", response.Code) }
}

func TestPortfolioPreviewRejectsOversizedAndInvalidContent(t *testing.T) {
	handler := NewRouter(Dependencies{
		Authorization: auth.NewService(fakeAuthRepository{membership: auth.Membership{Role: "admin"}}),
		AllowDevelopmentIdentity: true,
	})
	tooLarge := httptest.NewRecorder()
	handler.ServeHTTP(tooLarge, previewRequest(strings.Repeat("x", maxPortfolioCSVBytes+1)))
	if tooLarge.Code != http.StatusRequestEntityTooLarge { t.Fatalf("expected 413, got %d", tooLarge.Code) }

	invalidType := previewRequest(testPortfolioCSV)
	invalidType.Header.Set("Content-Type", "application/json")
	invalidResult := httptest.NewRecorder()
	handler.ServeHTTP(invalidResult, invalidType)
	if invalidResult.Code != http.StatusUnsupportedMediaType { t.Fatalf("expected 415, got %d", invalidResult.Code) }

	badCSV := httptest.NewRecorder()
	handler.ServeHTTP(badCSV, previewRequest("name,propertyType\nHome,house\n"))
	if badCSV.Code != http.StatusBadRequest { t.Fatalf("expected 400, got %d", badCSV.Code) }
}

func TestPortfolioPreviewNotAvailableWithoutAuth(t *testing.T) {
	request := previewRequest(testPortfolioCSV)
	request.Header.Del("X-User-ID")
	response := httptest.NewRecorder()
	NewRouter(Dependencies{Authorization: auth.NewService(fakeAuthRepository{membership: auth.Membership{Role: "admin"}}), AllowDevelopmentIdentity: false}).ServeHTTP(response, request)
	if response.Code != http.StatusUnauthorized { t.Fatalf("expected 401, got %d", response.Code) }
}
