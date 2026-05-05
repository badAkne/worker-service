package rprocessor

import (
	"net/http"

	"github.com/gorilla/mux"
)

// registerHealthRoutes регистрирует health check endpoints.
func registerHealthRoutes(r *mux.Router) {
	// Простой health check — возвращает 200 OK
	reg(r, http.MethodGet, "/health", http.HandlerFunc(healthHandler))
}

// TODO: Добавить при необходимости:
// func registerPprofRoutes(r *mux.Router) { ... }
// func registerMetricsRoutes(r *mux.Router) { ... }

// healthHandler — простой health check handler.
func healthHandler(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte("OK"))
}

// notFoundHandler — обработчик 404 Not Found.
func notFoundHandler(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusNotFound)
	_, _ = w.Write([]byte(`{"error":"not found"}`))
}

// methodNotAllowedHandler — обработчик 405 Method Not Allowed.
func methodNotAllowedHandler(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusMethodNotAllowed)
	_, _ = w.Write([]byte(`{"error":"method not allowed"}`))
}
