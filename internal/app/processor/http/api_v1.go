package rprocessor

import (
	"net/http"

	"github.com/gorilla/mux"

	rhandler "github.com/badAkne/worker-service/internal/app/handler/http"
)

// registerV1Routes регистрирует API v1 routes.
func registerV1Routes(r *mux.Router, hExample rhandler.Example) {
	v1 := r.PathPrefix("/v1").Subrouter()

	// Example routes
	if hExample != nil {
		registerExampleRoutes(v1, hExample)
	}

	// TODO: Добавить регистрацию других handlers:
	// if hUser != nil {
	//     registerUserRoutes(v1, hUser)
	// }
}

// registerExampleRoutes регистрирует example endpoints.
//
// Endpoints:
//   - POST /v1/example — принимает JSON, возвращает обработанный результат
//   - GET  /v1/example — принимает query params, возвращает результат
func registerExampleRoutes(r *mux.Router, h rhandler.Example) {
	reg(r, http.MethodPost, "/example", http.HandlerFunc(h.Post))
	reg(r, http.MethodGet, "/example", http.HandlerFunc(h.Get))
}

// TODO: Добавить функции регистрации для других handlers:
//
// func registerUserRoutes(r *mux.Router, h rhandler.User) {
//     reg(r, http.MethodPost, "/users", http.HandlerFunc(h.Create))
//     reg(r, http.MethodGet, "/users/{id}", http.HandlerFunc(h.Get))
//     reg(r, http.MethodGet, "/users", http.HandlerFunc(h.List))
//     reg(r, http.MethodPut, "/users/{id}", http.HandlerFunc(h.Update))
//     reg(r, http.MethodDelete, "/users/{id}", http.HandlerFunc(h.Delete))
// }
