package mzerolog

import (
	"net/http"

	"github.com/rs/zerolog"
)

// Option — функция для настройки middleware.
type Option = func(m *middleware)

// WithLogger устанавливает кастомный zerolog.Logger для middleware.
func WithLogger(l zerolog.Logger) Option {
	return func(m *middleware) {
		m.log = l
	}
}

// WithSkipper устанавливает функцию-фильтр для пропуска логирования определённых запросов.
// Если skipper возвращает true, запрос не будет залоггирован.
func WithSkipper(skipper func(r *http.Request) bool) Option {
	return func(m *middleware) {
		if skipper != nil {
			m.fromOptions.skipper = skipper
		}
	}
}

// WithStringExtractorOnSuccess добавляет extractor для успешных запросов.
// Извлекает строковое значение из request и логирует его с указанным ключом.
func WithStringExtractorOnSuccess(key string, cb CallbackExtractorString) Option {
	return func(m *middleware) {
		m.fromOptions.extStrOnSuccess = append(
			m.fromOptions.extStrOnSuccess, newStringExtractor(key, cb))
	}
}

// WithStringExtractorOnFail добавляет extractor для неуспешных запросов.
func WithStringExtractorOnFail(key string, cb CallbackExtractorString) Option {
	return func(m *middleware) {
		m.fromOptions.extStrOnFail = append(
			m.fromOptions.extStrOnFail, newStringExtractor(key, cb))
	}
}

// WithAnyExtractorOnSuccess добавляет extractor для успешных запросов.
// Извлекает любое значение из request и логирует его с указанным ключом.
func WithAnyExtractorOnSuccess(key string, cb CallbackExtractorAny) Option {
	return func(m *middleware) {
		m.fromOptions.extAnyOnSuccess = append(
			m.fromOptions.extAnyOnSuccess, newAnyExtractor(key, cb))
	}
}

// WithAnyExtractorOnFail добавляет extractor для неуспешных запросов.
func WithAnyExtractorOnFail(key string, cb CallbackExtractorAny) Option {
	return func(m *middleware) {
		m.fromOptions.extAnyOnFail = append(
			m.fromOptions.extAnyOnFail, newAnyExtractor(key, cb))
	}
}
