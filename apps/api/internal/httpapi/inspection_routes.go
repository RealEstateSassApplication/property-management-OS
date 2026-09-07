package httpapi

import (
	"net/http"

	"github.com/RealEstateSassApplication/property-management-OS/apps/api/internal/auth"
	"github.com/RealEstateSassApplication/property-management-OS/apps/api/internal/inspections"
)

// WithInspectionRoutes layers inspection endpoints over the existing API router
// without giving the inspection domain direct access to unrelated repositories.
func WithInspectionRoutes(base http.Handler, authentication *auth.Authenticator, authorization *auth.Service, allowDevelopmentIdentity bool, service *inspections.Service) http.Handler {
	mux := http.NewServeMux()
	h := inspectionHandler{service: service}
	protect := func(permission auth.Permission, handler http.HandlerFunc) http.Handler {
		return requirePermissionWithAuthentication(allowDevelopmentIdentity, authentication, authorization, permission, handler)
	}

	mux.Handle("GET /api/v1/inspections", protect(auth.ViewInspections, h.list))
	mux.Handle("POST /api/v1/inspections", protect(auth.ManageInspections, h.create))
	mux.Handle("GET /api/v1/inspections/{inspectionID}", protect(auth.ViewInspections, h.get))
	mux.Handle("GET /api/v1/inspections/{inspectionID}/items", protect(auth.ViewInspections, h.listItems))
	mux.Handle("POST /api/v1/inspections/{inspectionID}/items", protect(auth.ManageInspections, h.createItem))
	mux.Handle("POST /api/v1/inspections/{inspectionID}/complete", protect(auth.ManageInspections, h.complete))
	mux.Handle("POST /api/v1/inspections/{inspectionID}/acknowledge", protect(auth.AcknowledgeInspections, h.acknowledge))
	mux.Handle("/", base)
	return mux
}
