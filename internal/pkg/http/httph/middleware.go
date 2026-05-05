package httph

import (
	"net/http"
	"slices"
)

// Middleware — алиас для middleware функции HTTP.
type Middleware = func(handler http.Handler) http.Handler

// MergeMiddlewares возвращает один handler, который оборачивает переданный 'handler'
// всеми переданными 'middlewares' по очереди.
//
// Если middlewares пустой, возвращается переданный handler.
//
// ВНИМАНИЕ!
// Нет проверок на nil. Фильтруйте middlewares/handler заранее при необходимости.
// См.: FilterNilMiddlewares(), FilterNilHandlers().
func MergeMiddlewares(middlewares []Middleware, handler http.Handler) http.Handler {
	slices.Reverse(middlewares)
	for _, middleware := range middlewares {
		handler = middleware(handler)
	}

	return handler
}

// FilterNilMiddlewares фильтрует переданные 'middlewares', возвращая только не-nil объекты.
func FilterNilMiddlewares(middlewares []Middleware) []Middleware {
	return slices.DeleteFunc(middlewares, func(m Middleware) bool {
		return m == nil
	})
}

// FilterNilHandlers фильтрует переданные 'handlers', возвращая только не-nil объекты.
func FilterNilHandlers(handlers []http.Handler) []http.Handler {
	return slices.DeleteFunc(handlers, func(h http.Handler) bool {
		return h == nil
	})
}
