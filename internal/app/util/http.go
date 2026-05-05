package util

import (
	"net/http"
	"strings"
)

// IsFilteredHttpRoute проверяет, должен ли HTTP запрос быть отфильтрован
// (не логироваться). Возвращает true для служебных endpoints:
// health, debug, metric.
func IsFilteredHttpRoute(r *http.Request) bool {
	path := r.RequestURI
	return strings.Contains(path, "health") ||
		strings.Contains(path, "debug") ||
		strings.Contains(path, "metric")
}
