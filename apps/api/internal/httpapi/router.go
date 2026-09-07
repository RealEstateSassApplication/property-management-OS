package httpapi

import (
	"net/http"
	"time"

	"github.com/RealEstateSassApplication/property-management-OS/apps/api/internal/auth"
	"github.com/RealEstateSassApplication/property-management-OS/apps/api/internal/leases"
	"github.com/RealEstateSassApplication/property-management-OS/apps/api/internal/properties"
	"github.com/RealEstateSassApplication/property-management-OS/apps/api/internal/tenancies"
	"github.com/RealEstateSassApplication/property-management-OS/apps/api/internal/tenants"
	"github.com/RealEstateSassApplication/property-management-OS/apps/api/internal/units"
)

type Dependencies struct {
	Authorization            *auth.Service
	Properties               *properties.Service
	Units                    *units.Service
	Tenants                  *tenants.Service
	Tenancies                *tenancies.Service
	Leases                   *leases.Service
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

	protect := func(permission auth.Permission, handler http.HandlerFunc) http.Handler {
		return requirePermission(deps.AllowDevelopmentIdentity, deps.Authorization, permission, handler)
	}

	if deps.Properties != nil {
		h := propertyHandler{service: deps.Properties}
		r.mux.Handle("GET /api/v1/properties", protect(auth.ViewPortfolio, h.list))
		r.mux.Handle("POST /api/v1/properties", protect(auth.ManagePortfolio, h.create))
		r.mux.Handle("GET /api/v1/properties/{propertyID}", protect(auth.ViewPortfolio, h.get))
		r.mux.Handle("PATCH /api/v1/properties/{propertyID}", protect(auth.ManagePortfolio, h.update))
	}

	if deps.Units != nil {
		h := unitHandler{service: deps.Units}
		r.mux.Handle("GET /api/v1/properties/{propertyID}/units", protect(auth.ViewPortfolio, h.listByProperty))
		r.mux.Handle("POST /api/v1/properties/{propertyID}/units", protect(auth.ManagePortfolio, h.create))
		r.mux.Handle("GET /api/v1/units/{unitID}", protect(auth.ViewPortfolio, h.get))
		r.mux.Handle("PATCH /api/v1/units/{unitID}", protect(auth.ManagePortfolio, h.update))
	}

	if deps.Tenants != nil {
		h := tenantHandler{service: deps.Tenants}
		r.mux.Handle("GET /api/v1/tenants", protect(auth.ViewPeople, h.list))
		r.mux.Handle("POST /api/v1/tenants", protect(auth.ManagePeople, h.create))
		r.mux.Handle("GET /api/v1/tenants/{tenantID}", protect(auth.ViewPeople, h.get))
		r.mux.Handle("PATCH /api/v1/tenants/{tenantID}", protect(auth.ManagePeople, h.update))
	}

	if deps.Tenancies != nil {
		h := tenancyHandler{service: deps.Tenancies}
		r.mux.Handle("GET /api/v1/tenancies", protect(auth.ViewLeases, h.list))
		r.mux.Handle("POST /api/v1/tenancies", protect(auth.ManageLeases, h.create))
		r.mux.Handle("GET /api/v1/tenancies/{tenancyID}", protect(auth.ViewLeases, h.get))
		r.mux.Handle("PATCH /api/v1/tenancies/{tenancyID}", protect(auth.ManageLeases, h.update))
	}

	if deps.Leases != nil {
		h := leaseHandler{service: deps.Leases}
		r.mux.Handle("GET /api/v1/leases", protect(auth.ViewLeases, h.list))
		r.mux.Handle("POST /api/v1/leases", protect(auth.ManageLeases, h.create))
		r.mux.Handle("GET /api/v1/leases/{leaseID}", protect(auth.ViewLeases, h.get))
		r.mux.Handle("PATCH /api/v1/leases/{leaseID}", protect(auth.ManageLeases, h.update))
	}
}

func (r *Router) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	r.mux.ServeHTTP(w, req)
}
