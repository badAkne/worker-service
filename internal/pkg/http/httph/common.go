package httph

import (
	"io"
	"net/http"
	"net/textproto"
	"strings"

	json "github.com/goccy/go-json"
)

// HTTP заголовки.
const (
	HeaderContentType = "Content-Type"
	HeaderServer      = "Server"
)

// MIME типы.
const (
	MIMEApplicationJSON            = "application/json"
	MIMEApplicationJSONCharsetUTF8 = "application/json; charset=utf-8"
	MIMETextPlainCharsetUTF8       = "text/plain; charset=utf-8"
)

////////////////////////////////////////////////////////////////////////////////
///// HTTP Request headers methods /////////////////////////////////////////////
////////////////////////////////////////////////////////////////////////////////

// Header возвращает значение HTTP заголовка.
func Header(r *http.Request, key string) string {
	return HeaderWithDefault(r, key, "")
}

// HeaderWithDefault возвращает значение HTTP заголовка или default.
func HeaderWithDefault(r *http.Request, key, defaultValue string) string {
	if r != nil && r.Header != nil && key != "" {
		if value := r.Header.Get(key); value != "" {
			return value
		}
	}
	return defaultValue
}

// HeaderContain проверяет содержит ли заголовок указанное значение.
func HeaderContain(r *http.Request, key, value string) bool {
	if value == "" {
		return false
	}
	if gotValue := Header(r, key); gotValue == "" {
		return false
	} else {
		return strings.Contains(gotValue, value)
	}
}

// HeadersMerge объединяет HTTP заголовки.
func HeadersMerge(a, b http.Header, checkDuplicates bool) http.Header {
	if len(b) == 0 {
		return a
	}

	for key, bValues := range b {
		key = textproto.CanonicalMIMEHeaderKey(key)

		aValues := a[key]
		if len(aValues) == 0 {
			a[key] = bValues
			continue
		}

		aValues = append(aValues, bValues...)

		if checkDuplicates {
			aValues = distinctStrings(aValues)
		}

		a[key] = aValues
	}

	return a
}

////////////////////////////////////////////////////////////////////////////////
///// HTTP Response generators /////////////////////////////////////////////////
////////////////////////////////////////////////////////////////////////////////

// SendEncoded кодирует объект в JSON и отправляет как HTTP ответ.
func SendEncoded(w http.ResponseWriter, _ *http.Request, statusCode int, obj any) {
	w.Header().Set(HeaderContentType, MIMEApplicationJSONCharsetUTF8)
	w.WriteHeader(statusCode)

	if obj != nil {
		_ = json.NewEncoder(w).Encode(obj)
	}
}

// SendString отправляет строку как HTTP ответ.
func SendString(w http.ResponseWriter, statusCode int, data string) {
	SendRaw(w, statusCode, MIMETextPlainCharsetUTF8, []byte(data))
}

// SendEmpty отправляет пустой HTTP ответ.
func SendEmpty(w http.ResponseWriter, statusCode int) {
	SendRaw(w, statusCode, "", nil)
}

// SendRaw отправляет сырые данные как HTTP ответ.
func SendRaw(w http.ResponseWriter, statusCode int, mimeType string, data []byte) {
	if mimeType != "" {
		w.Header().Set(HeaderContentType, mimeType)
	}

	w.WriteHeader(statusCode)
	if len(data) != 0 {
		_, _ = w.Write(data)
	}
}

// SendStream отправляет данные из io.Reader как HTTP ответ.
func SendStream(w http.ResponseWriter, statusCode int, mimeType string, stream io.Reader) {
	SendRaw(w, statusCode, mimeType, nil)
	_, _ = io.Copy(w, stream)
}

////////////////////////////////////////////////////////////////////////////////
///// INTERNAL /////////////////////////////////////////////////////////////////
////////////////////////////////////////////////////////////////////////////////

// distinctStrings убирает дубликаты из слайса строк.
func distinctStrings(s []string) []string {
	seen := make(map[string]struct{}, len(s))
	result := make([]string, 0, len(s))
	for _, v := range s {
		if _, ok := seen[v]; !ok {
			seen[v] = struct{}{}
			result = append(result, v)
		}
	}
	return result
}
