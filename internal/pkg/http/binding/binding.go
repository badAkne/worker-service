package binding

import (
	"net/http"
)

// MIME типы для HTTP запросов.
const (
	MIMEJSON     = "application/json"
	MIMEPOSTForm = "application/x-www-form-urlencoded"
)

// Binding — интерфейс для парсинга данных из HTTP запроса.
type Binding interface {
	Name() string
	Bind(r *http.Request, obj any) error
}

// BindingBody добавляет метод BindBody к Binding.
// Позволяет парсить данные из []byte вместо req.Body.
type BindingBody interface {
	Binding
	BindBody(body []byte, obj any) error
}

// StructValidator — интерфейс для валидации структур.
type StructValidator interface {
	// ValidateStruct валидирует структуру.
	// Если obj не структура — возвращает nil.
	ValidateStruct(obj any) error

	// Engine возвращает underlying validator engine.
	Engine() any
}

// Validator — глобальный валидатор.
var Validator StructValidator = &defaultValidator{}

// Доступные bindings.
var (
	bJSON         = jsonBinding{}
	bQuery        = queryBinding{}
	bOnlyValidate = onlyValidateBinding{}
)

// Default возвращает Binding на основе HTTP метода и Content-Type.
func Default(method, contentType string) Binding {
	if method == http.MethodGet {
		return bQuery
	}
	if contentType == MIMEJSON {
		return bJSON
	}
	return bQuery
}

func validate(obj any) error {
	if Validator == nil {
		return nil
	}
	return Validator.ValidateStruct(obj)
}
