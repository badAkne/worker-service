package binding

import (
	"errors"
	"net/http"

	"github.com/go-playground/validator/v10"

	"github.com/badAkne/worker-service/internal/pkg/http/httph"
	"github.com/badAkne/worker-service/internal/pkg/http/respondent"
)

// Ошибки binding.
//
//goland:noinspection GoErrorStringFormat
var (
	// ErrMalformedSource — некорректный формат данных (JSON syntax error и т.д.)
	ErrMalformedSource = errors.New("Binding: Malformed HTTP source")

	// ErrValidationFailed — ошибка валидации (используется для errors.Is)
	ErrValidationFailed = (*validationFailedError)(nil)
)

// ScanAndValidateJSON парсит JSON body и валидирует результат.
//
// Пример использования:
//
//	var req CreateUserRequest
//	if err := binding.ScanAndValidateJSON(r, &req); err != nil {
//	    httph.ErrorPrepare(r, err)
//	    return
//	}
func ScanAndValidateJSON(r *http.Request, to any) error {
	return scanAndValidate(r, to, bJSON)
}

// ScanAndValidateQuery парсит query параметры и валидирует результат.
//
// Пример использования:
//
//	var req ListUsersRequest
//	if err := binding.ScanAndValidateQuery(r, &req); err != nil {
//	    httph.ErrorPrepare(r, err)
//	    return
//	}
func ScanAndValidateQuery(r *http.Request, to any) error {
	return scanAndValidate(r, to, bQuery)
}

// OnlyValidate только валидирует структуру без парсинга.
func OnlyValidate(to any) error {
	return scanAndValidate(nil, to, bOnlyValidate)
}

func scanAndValidate(r *http.Request, to any, b Binding) error {
	err := b.Bind(r, to)
	if err == nil {
		return nil
	}

	var validationErr validator.ValidationErrors
	if errors.As(err, &validationErr) {
		return &validationFailedError{validationErr}
	}

	if r != nil {
		httph.ErrorApplyDetail(r, "Malformed HTTP request "+b.Name()+" source")
	}
	return ErrMalformedSource
}

// NewRespondentManifestExtractor создаёт extractor для respondent.
// Преобразует ошибки валидации в структурированный Manifest.
//
// Использование в makeErrorMiddleware:
//
//	rce.ExtractorFor(
//	    binding.ErrValidationFailed, true,
//	    binding.NewRespondentManifestExtractor(http.StatusBadRequest, 40001, "Bad request"))
func NewRespondentManifestExtractor(status, errorCode int, message string) respondent.ManifestExtractor {
	return func(err error) *respondent.Manifest {
		manifest := respondent.Manifest{
			Status:    status,
			ErrorCode: errorCode,
			Error:     message,
		}

		var errList validator.ValidationErrors
		var typedErr *validationFailedError

		switch {
		case errors.As(err, &errList):
			// errList уже заполнен
		case errors.As(err, &typedErr):
			errList = typedErr.originalErr
		default:
			return nil
		}

		manifest.ErrorDetails = make([]string, len(errList))
		for i := range errList {
			manifest.ErrorDetails[i] = errList[i].Error()
		}

		return &manifest
	}
}

// validationFailedError — обёртка над validator.ValidationErrors.
type validationFailedError struct {
	originalErr validator.ValidationErrors
}

func (e *validationFailedError) Error() string {
	return "Binding: Validation failed"
}

func (e *validationFailedError) Is(other error) bool {
	var errValidationFailed *validationFailedError
	return errors.As(other, &errValidationFailed)
}
