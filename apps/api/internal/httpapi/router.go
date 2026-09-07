package httpapi

import (
	"net/http"
	"time"

	"github.com/RealEstateSassApplication/property-management-OS/apps/api/internal/auth"
	"github.com/RealEstateSassApplication/property-management-OS/apps/api/internal/documents"
	"github.com/RealEstateSassApplication/property-management-OS/apps/api/internal/leases"
	"github.com/RealEstateSassApplication/property-management-OS/apps/api/internal/maintenance"
	"github.com/RealEstateSassApplication/property-management-OS/apps/api/internal/owners"
	"github.com/RealEstateSassApplication/property-management-OS/apps/api/internal/properties"
	"github.com/RealEstateSassApplication/property-management-OS/apps/api/internal/rent"
	"github.com/RealEstateSassApplication/property-management-OS/apps/api/internal/tenancies"
	"github.com/RealEstateSassApplication/property-management-OS/apps/api/internal/tenants"
	"github.com/RealEstateSassApplication/property-management-OS/apps/api/internal/units"
)

type Dependencies struct {
	Authentication           *auth.Authenticator
	Authorization            *auth.Service
	Properties               *properties.Service
	Units                    *units.Service
	Tenants                  *tenants.Service
	Tenancies                *tenancies.Service
	Leases                   *leases.Service
	Owners                    *owners.Service
	Rent                      *rent.Service
	Maintenance               *maintenance.Service
	Documents                 *documents.Service
	AllowDevelopmentIdentity bool
}

type Router struct{ mux *http.ServeMux }

func NewRouter(deps Dependencies) http.Handler {
	r := &Router{mux: http.NewServeMux()}
	r.routes(deps)
	return r
}

func (r *Router) routes(deps Dependencies) {
	r.mux.HandleFunc("GET /healthz", func(w http.ResponseWriter, _ *http.Request) {
		writeJSON(w, http.StatusOK, map[string]any{"status": "ok", "service": "property-management-os-api", "timestamp": time.Now().UTC().Format(time.RFC3339)})
	})
	r.mux.HandleFunc("GET /api/v1", func(w http.ResponseWriter, _ *http.Request) {
		writeJSON(w, http.StatusOK, map[string]string{"name": "Property Management OS API", "version": "v1"})
	})
	protect := func(permission auth.Permission, handler http.HandlerFunc) http.Handler {
		return requirePermissionWithAuthentication(deps.AllowDevelopmentIdentity, deps.Authentication, deps.Authorization, permission, handler)
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
	if deps.Owners != nil {
		h := ownerHandler{service: deps.Owners}
		r.mux.Handle("GET /api/v1/owners", protect(auth.ViewOwners, h.list))
		r.mux.Handle("POST /api/v1/owners", protect(auth.ManageOwners, h.create))
		r.mux.Handle("GET /api/v1/owners/{ownerID}", protect(auth.ViewOwners, h.get))
		r.mux.Handle("GET /api/v1/ownership-interests", protect(auth.ViewOwners, h.listInterests))
		r.mux.Handle("POST /api/v1/ownership-interests", protect(auth.ManageOwners, h.createInterest))
	}
	if deps.Rent != nil {
		h := rentHandler{service: deps.Rent}
		r.mux.Handle("GET /api/v1/rent/obligations", protect(auth.ViewRent, h.listObligations))
		r.mux.Handle("POST /api/v1/rent/obligations", protect(auth.ManageRent, h.createObligation))
		r.mux.Handle("GET /api/v1/rent/payments", protect(auth.ViewRent, h.listPayments))
		r.mux.Handle("POST /api/v1/rent/payments", protect(auth.ManageRent, h.createPayment))
		r.mux.Handle("POST /api/v1/rent/allocations", protect(auth.ManageRent, h.createAllocation))
	}
	if deps.Maintenance != nil {
		h := maintenanceHandler{service: deps.Maintenance}
		r.mux.Handle("GET /api/v1/maintenance/vendors", protect(auth.ViewMaintenance, h.listVendors))
		r.mux.Handle("POST /api/v1/maintenance/vendors", protect(auth.ManageMaintenanceVendors, h.createVendor))
		r.mux.Handle("GET /api/v1/maintenance/requests", protect(auth.ViewMaintenance, h.listRequests))
		r.mux.Handle("POST /api/v1/maintenance/requests", protect(auth.ManageMaintenance, h.createRequest))
		r.mux.Handle("PATCH /api/v1/maintenance/requests/{requestID}/status", protect(auth.ManageMaintenance, h.updateRequestStatus))
		r.mux.Handle("GET /api/v1/maintenance/work-orders", protect(auth.ViewMaintenance, h.listWorkOrders))
		r.mux.Handle("POST /api/v1/maintenance/work-orders", protect(auth.ManageMaintenance, h.createWorkOrder))
		r.mux.Handle("PATCH /api/v1/maintenance/work-orders/{workOrderID}/status", protect(auth.ManageMaintenance, h.updateWorkOrderStatus))
		r.mux.Handle("GET /api/v1/maintenance/quotes", protect(auth.ViewMaintenance, h.listQuotes))
		r.mux.Handle("POST /api/v1/maintenance/quotes", protect(auth.ManageMaintenance, h.createQuote))
		r.mux.Handle("POST /api/v1/maintenance/quotes/{quoteID}/decision", protect(auth.ApproveMaintenanceCosts, h.decideQuote))
		r.mux.Handle("GET /api/v1/maintenance/evidence", protect(auth.ViewMaintenance, h.listEvidence))
		r.mux.Handle("POST /api/v1/maintenance/evidence", protect(auth.ManageMaintenance, h.createEvidence))
	}
	if deps.Documents != nil {
		h := documentHandler{service: deps.Documents}
		r.mux.Handle("GET /api/v1/documents", protect(auth.ViewDocuments, h.list))
		r.mux.Handle("POST /api/v1/documents/uploads", protect(auth.ManageDocuments, h.initiateUpload))
		r.mux.Handle("POST /api/v1/documents/{documentID}/complete", protect(auth.ManageDocuments, h.completeUpload))
		r.mux.Handle("GET /api/v1/documents/{documentID}/download", protect(auth.ViewDocuments, h.download))
		r.mux.Handle("DELETE /api/v1/documents/{documentID}", protect(auth.ManageDocuments, h.delete))
	}
}
func (r *Router) ServeHTTP(w http.ResponseWriter, req *http.Request) { r.mux.ServeHTTP(w, req) }
