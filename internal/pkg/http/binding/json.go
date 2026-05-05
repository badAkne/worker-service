package binding

import (
	"bytes"
	"errors"
	"io"
	"net/http"

	json "github.com/goccy/go-json"
)

type jsonBinding struct{}

func (jsonBinding) Name() string {
	return "JSON"
}

func (jsonBinding) Bind(req *http.Request, obj any) error {
	if req == nil || req.Body == nil {
		return errors.New("invalid request")
	}
	return decodeJSON(req.Body, obj)
}

func (jsonBinding) BindBody(body []byte, obj any) error {
	return decodeJSON(bytes.NewReader(body), obj)
}

func decodeJSON(r io.Reader, obj any) error {
	decoder := json.NewDecoder(r)
	if err := decoder.Decode(obj); err != nil && !errors.Is(err, io.EOF) {
		return err
	}
	return validate(obj)
}
