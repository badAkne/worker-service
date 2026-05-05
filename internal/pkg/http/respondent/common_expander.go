package respondent

import (
	"errors"
	"net/http"
	"reflect"
)

// CommonExpander — реализация Expander интерфейса.
// Представляет собой базу правил для преобразования ошибок в Manifest.
type CommonExpander struct {
	// Логика:
	// 1. Direct — ошибки, которые можно сравнивать напрямую (==)
	// 2. Deep — ошибки, которые сравниваются через errors.Is()
	// 3. Check — ошибки, проверяемые кастомной функцией
	// 4. Fallback — вызывается если ничего не подошло

	direct   map[error]ManifestExtractor
	deep     []commonExpanderDeepPair
	check    []commonExpanderCheckPair
	fallback ManifestExtractor
}

// ManifestExtractor — функция для создания Manifest из ошибки.
type ManifestExtractor = func(err error) *Manifest

type commonExpanderDeepPair struct {
	Pattern   error
	Extractor ManifestExtractor
}

type commonExpanderCheckPair struct {
	Checker   func(err error) bool
	Extractor ManifestExtractor
}

// Expand преобразует ошибку в Manifest.
func (ce *CommonExpander) Expand(err error) *Manifest {
	if ce == nil || err == nil {
		return nil
	}

	var manifest *Manifest

	// Пробуем прямое сравнение
	if isGoHashableObject(reflect.TypeOf(err).Kind()) {
		if extractor := ce.direct[err]; extractor != nil {
			manifest = extractor(err)
		}
	}

	// Пробуем глубокое сравнение через errors.Is()
	for i, n := 0, len(ce.deep); i < n && manifest == nil; i++ {
		if errors.Is(err, ce.deep[i].Pattern) {
			manifest = ce.deep[i].Extractor(err)
			break
		}
	}

	// Пробуем кастомные чекеры
	for i, n := 0, len(ce.check); i < n && manifest == nil; i++ {
		if ce.check[i].Checker(err) {
			manifest = ce.check[i].Extractor(err)
			break
		}
	}

	// Fallback
	if ce.fallback != nil && manifest == nil {
		manifest = ce.fallback(err)
	}

	return manifest
}

// ExtractorFor регистрирует extractor для указанной ошибки.
// deep=true — использовать errors.Is() для сравнения.
func (ce *CommonExpander) ExtractorFor(err error, deep bool, cb ManifestExtractor) *CommonExpander {
	if ce == nil || err == nil || cb == nil {
		return ce
	}

	if deep {
		ce.deep = append(ce.deep, commonExpanderDeepPair{err, cb})
	} else {
		if ce.direct == nil {
			ce.direct = make(map[error]ManifestExtractor)
		}
		ce.direct[err] = cb
	}

	return ce
}

// ExtractorByChecker регистрирует extractor с кастомной функцией проверки.
func (ce *CommonExpander) ExtractorByChecker(checker func(err error) bool, cb ManifestExtractor) *CommonExpander {
	if ce != nil && checker != nil && cb != nil {
		ce.check = append(ce.check, commonExpanderCheckPair{checker, cb})
	}
	return ce
}

// ManifestFor регистрирует готовый Manifest для указанной ошибки.
func (ce *CommonExpander) ManifestFor(err error, deep bool, manifest *Manifest) *CommonExpander {
	return ce.ExtractorFor(err, deep, ce.createSimpleExtractor(manifest))
}

// FallbackExtractor устанавливает fallback extractor.
func (ce *CommonExpander) FallbackExtractor(cb ManifestExtractor) *CommonExpander {
	if ce != nil && cb != nil {
		ce.fallback = cb
	}
	return ce
}

func (*CommonExpander) createSimpleExtractor(manifest *Manifest) ManifestExtractor {
	if manifest != nil {
		return func(_ error) *Manifest { return manifest }
	}
	return nil
}

// NewCommonExpander создаёт новый CommonExpander с дефолтным fallback.
func NewCommonExpander() *CommonExpander {
	return &CommonExpander{
		direct: make(map[error]ManifestExtractor),
		deep:   make([]commonExpanderDeepPair, 0, 32),
		check:  make([]commonExpanderCheckPair, 0, 32),
		fallback: func(err error) *Manifest {
			return &Manifest{
				Status:      http.StatusInternalServerError,
				Error:       "Unrecoverable internal server error",
				ErrorCode:   http.StatusInternalServerError * 100,
				ErrorDetail: err.Error(),
			}
		},
	}
}

////////////////////////////////////////////////////////////////////////////////
///// КОРОТКИЕ АЛИАСЫ //////////////////////////////////////////////////////////
////////////////////////////////////////////////////////////////////////////////

// WithoutDetail — алиас для ManifestFor без детализации.
func (ce *CommonExpander) WithoutDetail(err error, status, errorCode int, msg string) *CommonExpander {
	return ce.ManifestFor(err, false, &Manifest{
		Status: status, ErrorCode: errorCode, Error: msg,
	})
}

// WithDetail — алиас для ManifestFor с одной строкой детализации.
func (ce *CommonExpander) WithDetail(err error, status, errorCode int, msg, reason string) *CommonExpander {
	return ce.ManifestFor(err, false, &Manifest{
		Status: status, ErrorCode: errorCode, Error: msg, ErrorDetail: reason,
	})
}

// WithDetails — алиас для ManifestFor с массивом детализации.
func (ce *CommonExpander) WithDetails(err error, status, errorCode int, msg string, reasons []string) *CommonExpander {
	return ce.ManifestFor(err, false, &Manifest{
		Status: status, ErrorCode: errorCode, Error: msg, ErrorDetails: reasons,
	})
}

// WithCustomFillers — алиас для ManifestFor с кастомными fillers.
func (ce *CommonExpander) WithCustomFillers(err error, status, errorCode int, msg string, ext ...ManifestCustomFiller) *CommonExpander {
	return ce.ManifestFor(err, false, &Manifest{
		Status: status, ErrorCode: errorCode, Error: msg, customFillers: ext,
	})
}

// DeepWithoutDetail — как WithoutDetail, но с глубоким сравнением (errors.Is).
func (ce *CommonExpander) DeepWithoutDetail(err error, status, errorCode int, msg string) *CommonExpander {
	return ce.ManifestFor(err, true, &Manifest{
		Status: status, ErrorCode: errorCode, Error: msg,
	})
}

// DeepWithDetail — как WithDetail, но с глубоким сравнением (errors.Is).
func (ce *CommonExpander) DeepWithDetail(err error, status, errorCode int, msg, reason string) *CommonExpander {
	return ce.ManifestFor(err, true, &Manifest{
		Status: status, ErrorCode: errorCode, Error: msg, ErrorDetail: reason,
	})
}

// DeepWithDetails — как WithDetails, но с глубоким сравнением (errors.Is).
func (ce *CommonExpander) DeepWithDetails(err error, status, errorCode int, msg string, reasons []string) *CommonExpander {
	return ce.ManifestFor(err, true, &Manifest{
		Status: status, ErrorCode: errorCode, Error: msg, ErrorDetails: reasons,
	})
}

// DeepWithCustomFillers — как WithCustomFillers, но с глубоким сравнением (errors.Is).
func (ce *CommonExpander) DeepWithCustomFillers(err error, status, errorCode int, msg string, ext ...ManifestCustomFiller) *CommonExpander {
	return ce.ManifestFor(err, true, &Manifest{
		Status: status, ErrorCode: errorCode, Error: msg, customFillers: ext,
	})
}

////////////////////////////////////////////////////////////////////////////////
///// INTERNAL /////////////////////////////////////////////////////////////////
////////////////////////////////////////////////////////////////////////////////

func isGoHashableObject(kind reflect.Kind) bool {
	switch kind {
	case reflect.Array, reflect.Slice, reflect.Map, reflect.Chan, reflect.Func:
		return false
	default:
		return true
	}
}
