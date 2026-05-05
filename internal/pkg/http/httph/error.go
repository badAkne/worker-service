package httph

import (
	"context"
	"net/http"
)

type _ContextKeyError struct{}

type _ContextValueError struct {
	err       error
	detail    string
	isHandled bool
}

// ErrorPrepare подготавливает request для хранения ошибки в контексте.
func ErrorPrepare(r *http.Request) *http.Request {
	return r.WithContext(errorPrepare(r.Context()))
}

// ErrorGet возвращает ошибку из контекста request.
func ErrorGet(r *http.Request) error {
	return errorGet(r.Context())
}

// ErrorGetDetail возвращает детали ошибки из контекста request.
func ErrorGetDetail(r *http.Request) string {
	return errorGetDetail(r.Context())
}

// ErrorGetByContext возвращает ошибку из контекста.
func ErrorGetByContext(ctx context.Context) error {
	return errorGet(ctx)
}

// ErrorGetDetailByContext возвращает детали ошибки из контекста.
func ErrorGetDetailByContext(ctx context.Context) string {
	return errorGetDetail(ctx)
}

// ErrorTryAcquireHandling пытается захватить обработку ошибки.
// Возвращает true, если ошибка ещё не была обработана.
func ErrorTryAcquireHandling(r *http.Request) bool {
	return errorTryAcquireHandling(r.Context())
}

// ErrorTryAcquireHandlingByContext — версия с context.
func ErrorTryAcquireHandlingByContext(ctx context.Context) bool {
	return errorTryAcquireHandling(ctx)
}

// ErrorApply устанавливает ошибку в контекст request.
func ErrorApply(r *http.Request, err error) {
	errorApply(r.Context(), err)
}

// ErrorApplyDetail устанавливает детали ошибки в контекст request.
func ErrorApplyDetail(r *http.Request, detail string) {
	errorApplyDetail(r.Context(), detail)
}

// NewErrorMiddleware создаёт middleware, который подготавливает контекст для хранения ошибок.
// Должен быть добавлен первым в цепочку middleware.
func NewErrorMiddleware() Middleware {
	return func(handler http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			handler.ServeHTTP(w, ErrorPrepare(r))
		})
	}
}

////////////////////////////////////////////////////////////////////////////////
///// PRIVATE METHODS //////////////////////////////////////////////////////////
////////////////////////////////////////////////////////////////////////////////

func errorPrepare(ctx context.Context) context.Context {
	errCtx := new(_ContextValueError)
	return context.WithValue(ctx, _ContextKeyError{}, errCtx)
}

func errorGet(ctx context.Context) error {
	errV, _ := ctx.Value(_ContextKeyError{}).(*_ContextValueError)
	if errV != nil {
		return errV.err
	}
	return nil
}

func errorGetDetail(ctx context.Context) string {
	errV, _ := ctx.Value(_ContextKeyError{}).(*_ContextValueError)
	if errV != nil {
		return errV.detail
	}
	return ""
}

func errorApply(ctx context.Context, err error) {
	errV, _ := ctx.Value(_ContextKeyError{}).(*_ContextValueError)
	if errV != nil {
		errV.err = err
	}
}

func errorApplyDetail(ctx context.Context, detail string) {
	errV, _ := ctx.Value(_ContextKeyError{}).(*_ContextValueError)
	if errV != nil {
		errV.detail = detail
	}
}

// errorTryAcquireHandling пытается установить флаг isHandled в true,
// если он еще не установлен. Если успешно, возвращает true, уведомляя,
// что текущий обработчик ошибок может обработать ошибку.
// Иначе (если уже флаг был установлен в true), возвращает false,
// оповещая, что ошибка уже была обработана и обработчик ошибок НЕ ДОЛЖЕН
// делать ничего.
func errorTryAcquireHandling(ctx context.Context) bool {
	errV, _ := ctx.Value(_ContextKeyError{}).(*_ContextValueError)
	if errV == nil || errV.isHandled {
		return false
	}
	errV.isHandled = true
	return true
}
