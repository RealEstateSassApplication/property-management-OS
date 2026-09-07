package httpapi

import (
	"encoding/json"
	"net/http"
	"time"
)

type Router struct {
	mux *http.ServeMux
}

func NewRouter() http.Handler {
	r := &Router{mux: http.NewServeMux()}
	r.routes()
	return r
}

func (r *Router) routes() {
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
}

func (r *Router) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	r.mux.ServeHTTP(w, req)
}

func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}
