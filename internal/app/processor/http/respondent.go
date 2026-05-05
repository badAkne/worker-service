package rprocessor

import (
	"net/http"

	"github.com/badAkne/worker-service/internal/pkg/http/binding"
	"github.com/badAkne/worker-service/internal/pkg/http/httph"
	"github.com/badAkne/worker-service/internal/pkg/http/respondent"
)

// makeErrorMiddleware создаёт middleware для обработки ошибок.
// Преобразует ошибки из контекста запроса в структурированные JSON ответы.
//
// Как добавить обработку новой ошибки:
//  1. Определите ошибку в entity пакете: var ErrUserNotFound = errors.New("user not found")
//  2. Добавьте маппинг ниже: WithoutDetail(entity.ErrUserNotFound, http.StatusNotFound, 40401, _40401)
//
// Коды ошибок (errorCode):
//   - 4XXNN — клиентские ошибки (4XX — HTTP статус, NN — порядковый номер)
//   - 5XXNN — серверные ошибки
func makeErrorMiddleware() httph.Middleware {
	// Сообщения об ошибках
	const (
		_40001 = "Некорректный запрос"
		_40002 = "Ошибка валидации"

		_50001  = "Внутренняя ошибка сервера"
		_50001D = "Повторите запрос позже. При повторении ошибки сообщите X-Request-ID"
	)

	// Хелпер для создания fallback extractor
	makeFallbackExtractor := func(status, errorCode int, message, detail string) respondent.ManifestExtractor {
		genericManifest := respondent.Manifest{
			Status:      status,
			Error:       message,
			ErrorCode:   errorCode,
			ErrorDetail: detail,
		}
		return func(_ error) *respondent.Manifest { return &genericManifest }
	}

	rce := respondent.NewCommonExpander().
		//
		// --------- HTTP 400: Bad Request --------- //
		//
		// Ошибка парсинга JSON/Query (синтаксическая ошибка)
		WithCustomFillers(binding.ErrMalformedSource, http.StatusBadRequest, 40001, _40001, respondent.CACF_AutoErrorDetail).
		// Ошибка валидации (binding:"required", binding:"email" и т.д.)
		ExtractorFor(
			binding.ErrValidationFailed, true,
			binding.NewRespondentManifestExtractor(http.StatusBadRequest, 40002, _40002)).
		//
		// TODO: Добавить бизнес-ошибки приложения:
		// WithoutDetail(entity.ErrInvalidInput, http.StatusBadRequest, 40003, "Некорректные данные").
		//
		// --------- HTTP 401: Unauthorized --------- //
		//
		// TODO: Добавить обработку ошибок авторизации:
		// WithoutDetail(entity.ErrUnauthorized, http.StatusUnauthorized, 40101, "Требуется авторизация").
		//
		// --------- HTTP 403: Forbidden --------- //
		//
		// TODO: Добавить обработку ошибок доступа:
		// WithoutDetail(entity.ErrForbidden, http.StatusForbidden, 40301, "Доступ запрещён").
		//
		// --------- HTTP 404: Not Found --------- //
		//
		// TODO: Добавить обработку ошибок "не найдено":
		// WithDetail(entity.ErrNotFound, http.StatusNotFound, 40401, "Не найдено", "Ресурс не существует").
		//
		// --------- HTTP 500: Internal Server Error --------- //
		//
		FallbackExtractor(makeFallbackExtractor(http.StatusInternalServerError, 50001, _50001, _50001D))

	return respondent.NewMiddleware(rce)
}
