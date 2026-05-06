package rprocessor

import (
	"net/http"
	"unsafe"

	"github.com/badAkne/worker-service/internal/pkg/http/httph"
	"github.com/gorilla/mux"
)

// reg — хелпер для регистрации route.
func reg(r *mux.Router, method, path string, handler http.Handler) {
	r.Methods(method).Path(path).Handler(handler)
}

// middlewaresToGorilla конвертирует []httph.Middleware в []mux.MiddlewareFunc.
//
// Эта функция существует, потому что gorilla использует type definition (type T T2)
// вместо type alias (type T = T2) для "mux.MiddlewareFunc".
//
// Поэтому типы "mux.MiddlewareFunc" и "httph.Middleware" не являются одинаковыми.
//
// Из-за этого нам нужна эта функция конвертации.
// Эти типы одинаковы на уровне байтов, поэтому безопасно использовать unsafe.
// Для гарантии, что прямая явная конвертация этих типов возможна,
// у нас есть compile check ниже.
// Пока он успешно компилируется, мы можем быть уверены, что типы одинаковы
// на уровне байтов.
func middlewaresToGorilla(m []httph.Middleware) []mux.MiddlewareFunc {
	_ = mux.MiddlewareFunc(httph.Middleware(nil)) // safety check
	//nolint:gosec // G103: безопасно, т.к. типы идентичны на уровне байтов
	return *(*[]mux.MiddlewareFunc)(unsafe.Pointer(&m))
}
