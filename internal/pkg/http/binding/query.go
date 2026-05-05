package binding

import (
	"net/http"

	"github.com/gorilla/schema"
)

// queryDecoder — глобальный декодер для query параметров.
var queryDecoder = func() *schema.Decoder {
	d := schema.NewDecoder()
	d.IgnoreUnknownKeys(true)
	d.SetAliasTag("form") // используем тот же тег что и для form
	return d
}()

type queryBinding struct{}

func (queryBinding) Name() string {
	return "URL-QUERY"
}

func (queryBinding) Bind(req *http.Request, obj any) error {
	if err := queryDecoder.Decode(obj, req.URL.Query()); err != nil {
		return err
	}
	return validate(obj)
}

// onlyValidateBinding — binding только для валидации (без парсинга).
type onlyValidateBinding struct{}

func (onlyValidateBinding) Name() string {
	return "OnlyValidate"
}

func (onlyValidateBinding) Bind(_ *http.Request, obj any) error {
	return validate(obj)
}
