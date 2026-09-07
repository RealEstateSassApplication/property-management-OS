package httpapi

import (
	"net/http"
	"time"

	"github.com/RealEstateSassApplication/property-management-OS/apps/api/internal/properties"
	"github.com/RealEstateSassApplication/property-management-OS/apps/api/internal/units"
)

type Dependencies struct {
	Properties               *properties.Service
	Units                    *units.Service
	AllowDevelopmentIdentity bool
}

type Router struct {
	mux *http.ServeMux
}

func NewRouter(deps Dependencies) http.Handler {
	r := &Router{mux: http.NewServeMux()}
	r.routes(deps)
	return r
}

func (r *Router) routes(deps Dependencies) {
	r.mux.HandleFunc("GET /healthz", func(w http.ResponseWriter, _ *http.Request) {
		writeJSON(w, http.StatusOK, map[string]any{
			"status":    "ok",
			"service":   "property-management-os-api",
			"timestamp": time.Now().UTC().Format(time.RFC3339),
		})
	})

	r.mux.HandleFunc("GET /api/v1", func(w http.ResponseWriter, _ *http.Request) {
		writeJSON(w, http.StatusOK, map[string]string{
			"name":    "Property Management OS API",
			"version": "v1",
		})
	})

	if deps.Properties != nil {
		propertiesHandler := propertyHandler{service: deps.Properties}
		r.mux.Handle("GET /api/v1/properties", requireOrganizationContext(deps.AllowDevelopmentIdentity, http.HandlerFunc(propertiesHandler.list)))
		r.mux.Handle("POST /api/v1/properties", requireOrganizationContext(deps.AllowDevelopmentIdentity, http.HandlerFunc(propertiesHandler.create)))
		r.mux.Handle("GET /api/v1/properties/{propertyID}", requireOrganizationContext(deps.AllowDevelopmentIdentity, http.HandlerFunc(propertiesHandler.get)))
		r.mux.Handle("PATCH /api/v1/properties/{propertyID}", requireOrganizationContext(deps.AllowDevelopmentIdentity, http.HandlerFunc(propertiesHandler.update)))
	}

	if deps.Units != nil {
		unitsHandler := unitHandler{service: deps.Units}
		r.mux.Handle("GET /api/v1/properties/{propertyID}/units", requireOrganizationContext(deps.AllowDevelopmentIdentity, http.HandlerFunc(unitsHandler.listByProperty)))
		r.mux.Handle("POST /api/v1/properties/{propertyID}/units", requireOrganizationContext(deps.AllowDevelopmentIdentity, http.HandlerFunc(unitsHandler.create)))
		r.mux.Handle("GET /api/v1/units/{unitID}", requireOrganizationContext(deps.AllowDevelopmentIdentity, http.HandlerFunc(unitsHandler.get)))
		r.mux.Handle("PATCH /api/v1/units/{unitID}", requireOrganizationContext(deps.AllowDevelopmentIdentity, http.HandlerFunc(unitsHandler.update)))
	}
}

func (r *Router) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	r.mux.ServeHTTP(w, req)
}
