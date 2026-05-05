package binding

import (
	"reflect"
	"sync"

	"github.com/go-playground/validator/v10"
)

type defaultValidator struct {
	once     sync.Once
	validate *validator.Validate
}

var _ StructValidator = &defaultValidator{}

// ValidateStruct валидирует структуру.
// Если obj не структура или указатель на структуру — возвращает nil.
func (v *defaultValidator) ValidateStruct(obj any) error {
	value := reflect.ValueOf(obj)
	valueType := value.Kind()
	if valueType == reflect.Ptr {
		valueType = value.Elem().Kind()
	}
	if valueType == reflect.Struct {
		v.lazyinit()
		if err := v.validate.Struct(obj); err != nil {
			return err
		}
	}
	return nil
}

// Engine возвращает underlying validator engine.
// Используйте для регистрации кастомных валидаторов.
func (v *defaultValidator) Engine() any {
	v.lazyinit()
	return v.validate
}

func (v *defaultValidator) lazyinit() {
	v.once.Do(func() {
		v.validate = validator.New()
		v.validate.SetTagName("binding")
	})
}

// RegisterCustomValidator регистрирует кастомный валидатор.
//
// Пример:
//
//	binding.RegisterCustomValidator("notempty", func(fl validator.FieldLevel) bool {
//	    return fl.Field().String() != ""
//	})
func RegisterCustomValidator(tag string, fn validator.Func) error {
	v, ok := Validator.Engine().(*validator.Validate)
	if !ok {
		return nil
	}
	return v.RegisterValidation(tag, fn)
}
