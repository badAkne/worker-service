package rhandler

import (
	"net/http"
)

// Example — интерфейс example handler.
// Определяет контракт для HTTP endpoints.
type Example interface {
	// Post обрабатывает POST /v1/example
	// Принимает JSON body, возвращает обработанный результат.
	Post(w http.ResponseWriter, r *http.Request)

	// Get обрабатывает GET /v1/example
	// Принимает query параметры, возвращает результат.
	Get(w http.ResponseWriter, r *http.Request)
}

// TODO: Добавить интерфейсы для других handlers:
//
// type User interface {
//     Create(w http.ResponseWriter, r *http.Request)
//     Get(w http.ResponseWriter, r *http.Request)
//     List(w http.ResponseWriter, r *http.Request)
//     Update(w http.ResponseWriter, r *http.Request)
//     Delete(w http.ResponseWriter, r *http.Request)
// }
