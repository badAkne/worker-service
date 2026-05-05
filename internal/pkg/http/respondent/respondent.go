package respondent

import (
	"errors"
	"net/http"

	"github.com/badAkne/worker-service/internal/pkg/http/httph"
)

// respondent — комбинация Replacer, Expander и Applicator,
// работающих вместе для преобразования ошибки в HTTP ответ.
type respondent struct {
	expander Expander

	fromOptions struct {
		replacer   Replacer
		applicator Applicator
	}
}

// HttpContext — контейнер для http.ResponseWriter и http.Request.
// Используется когда нужно передать оба в одном аргументе.
type HttpContext struct {
	W http.ResponseWriter
	R *http.Request
}

//goland:noinspection GoErrorStringFormat
var (
	// ErrEmptyReplacement возвращается когда Replacer возвращает nil.
	ErrEmptyReplacement = errors.New("Middleware.Respondent got unexpectedly empty replacement")

	// ErrEmptyManifest возвращается когда Expander возвращает nil.
	ErrEmptyManifest = errors.New("Middleware.Respondent got unexpectedly empty manifest")

	errBadExpander = errors.New("Middleware.Respondent is not initialized properly: Nil or incorrect Expander")
)

// Callback — основной метод Respondent.
// Преобразует ошибку в HTTP ответ через цепочку: Replacer → Expander → Applicator.
func (rp *respondent) Callback(ctx any, err error) {
	err = rp.fromOptions.replacer.Replace(err)
	if err == nil {
		return
	}

	manifest := rp.expander.Expand(err)
	if manifest == nil {
		return
	}

	rp.fromOptions.applicator.Apply(ctx, manifest)
}

// CallbackForHTTP — версия Callback с HTTP-совместимой сигнатурой.
func (rp *respondent) CallbackForHTTP(w http.ResponseWriter, r *http.Request, err error) {
	rp.Callback(HttpContext{w, r}, err)
}

// newRespondent создаёт новый Respondent объект.
func newRespondent(expander Expander, opts []Option) *respondent {
	m := respondent{}
	m.fromOptions.replacer = newNoOpReplacer()
	m.fromOptions.applicator = NewCommonApplicator()

	for i, n := 0, len(opts); i < n; i++ {
		if opts[i] != nil {
			opts[i](&m)
		}
	}

	if expander == nil {
		panic(errBadExpander)
	}

	m.expander = expander
	return &m
}

// NewMiddleware создаёт HTTP middleware для обработки ошибок.
// Использует ошибку из контекста запроса (httph.ErrorGet) и преобразует её в JSON ответ.
func NewMiddleware(expander Expander, opts ...Option) httph.Middleware {
	resp := newRespondent(expander, opts)
	return func(handler http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			handler.ServeHTTP(w, r)

			err := httph.ErrorGet(r)
			mayHandle := httph.ErrorTryAcquireHandling(r)

			if err != nil && mayHandle {
				resp.CallbackForHTTP(w, r, err)
			}
		})
	}
}
