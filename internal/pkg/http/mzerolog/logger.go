package mzerolog

import (
	"net/http"
	"strings"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"github.com/badAkne/worker-service/internal/pkg/http/httph"
)

// ErrorGet — локальный алиас для удобства.
var ErrorGet = httph.ErrorGet

type (
	// middleware предоставляет возможность создания
	// logging middleware на основе https://github.com/rs/zerolog .
	middleware struct {
		log zerolog.Logger

		fromOptions struct {
			extStrOnSuccess []extractorStr
			extAnyOnSuccess []extractorAny

			extStrOnFail []extractorStr
			extAnyOnFail []extractorAny

			skipper func(r *http.Request) bool
		}
	}

	// extractorStr — пара из callback'а для извлечения строки из http.Request
	// и ключа, с которым извлечённое значение будет залоггировано.
	extractorStr struct {
		key string
		ext CallbackExtractorString
	}

	// extractorAny — то же что extractorStr, но для любых значений.
	extractorAny struct {
		key string
		ext CallbackExtractorAny
	}

	// CallbackExtractorString — функция для извлечения строкового значения из http.Request.
	CallbackExtractorString = func(r *http.Request) string

	// CallbackExtractorAny — функция для извлечения любого значения из http.Request.
	CallbackExtractorAny = func(r *http.Request) any
)

// Callback реализует middleware. Возвращает http.Handler,
// который логирует HTTP запросы, оборачивая вызов следующего handler'а.
func (m *middleware) Callback(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		const (
			TailSuccess = " finished with no error"           // 23 chars
			TailFail    = " finished (or aborted) with error" // 33 chars
		)

		start := time.Now()

		next.ServeHTTP(w, r)
		err := ErrorGet(r)

		execTime := time.Since(start)

		// Мы НЕ должны фильтровать запрос раньше, поскольку фильтр может
		// полагаться на наличие или отсутствие ошибки. Таким образом,
		// мы должны позволить самому обработчику отработать.

		if m.fromOptions.skipper(r) {
			return
		}

		// Собираем сообщение: HTTP method (max=7), URI, результат.
		var mb strings.Builder
		mb.Grow(48 + len(r.RequestURI))
		mb.WriteString(r.Method)
		mb.WriteByte(' ')
		mb.WriteString(r.RequestURI)

		var ev *zerolog.Event

		var extString []extractorStr
		var extAny []extractorAny

		if completedWithNoError := err == nil; completedWithNoError {
			mb.WriteString(TailSuccess)
			ev = m.log.Debug()
			extString = m.fromOptions.extStrOnSuccess
			extAny = m.fromOptions.extAnyOnSuccess
		} else {
			mb.WriteString(TailFail)
			ev = m.log.Error()
			extString = m.fromOptions.extStrOnFail
			extAny = m.fromOptions.extAnyOnFail
		}

		m.applyExtractors(r, ev, extString, extAny)

		ev.Err(err)
		ev.Str("exec_time", execTime.String())
		ev.Str("client_ip", r.RemoteAddr)

		ev.Msg(mb.String())
	})
}

// applyExtractors вызывает extractors и добавляет извлечённые значения в event.
func (*middleware) applyExtractors(
	r *http.Request, ev *zerolog.Event,
	extractorsString []extractorStr, extractorsAny []extractorAny,
) {
	for i, n := 0, len(extractorsString); i < n; i++ {
		key := extractorsString[i].key
		valueString := extractorsString[i].ext(r)
		if valueString != "" {
			ev.Str(key, valueString)
		}
	}

	for i, n := 0, len(extractorsAny); i < n; i++ {
		key := extractorsAny[i].key
		valueAny := extractorsAny[i].ext(r)
		if valueAny != nil {
			ev.Any(key, valueAny)
		}
	}
}

////////////////////////////////////////////////////////////////////////////////
////////////////////////////////////////////////////////////////////////////////
////////////////////////////////////////////////////////////////////////////////

// NewMiddleware возвращает новый logging middleware для логирования
// входящих HTTP запросов. Использует https://github.com/rs/zerolog .
// Можно передать Option(s) для настройки поведения.
func NewMiddleware(opts ...Option) httph.Middleware {
	m := middleware{log: log.Logger}
	m.fromOptions.skipper = defaultSkipper

	for i, n := 0, len(opts); i < n; i++ {
		if opts[i] != nil {
			opts[i](&m)
		}
	}

	return m.Callback
}

// newStringExtractor — конструктор extractorStr.
func newStringExtractor(key string, cb CallbackExtractorString) extractorStr {
	return extractorStr{key, cb}
}

// newAnyExtractor — конструктор extractorAny.
func newAnyExtractor(key string, cb CallbackExtractorAny) extractorAny {
	return extractorAny{key, cb}
}

// defaultSkipper — функция-фильтр HTTP запросов, которые НЕ должны быть
// залоггированы. Всегда возвращает false, т.е. все запросы будут залоггированы.
func defaultSkipper(*http.Request) bool { return false }
