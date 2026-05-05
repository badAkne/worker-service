package respondent

import (
	"encoding/json"
	"net/http"

	"github.com/badAkne/worker-service/internal/pkg/http/httph"
)

// CommonApplicator — реализация Applicator интерфейса.
// Преобразует Manifest в JSON HTTP ответ.
type CommonApplicator struct{}

// Apply преобразует Manifest в JSON и отправляет как HTTP ответ.
func (*CommonApplicator) Apply(ctx any, manifest *Manifest) {
	//goland:noinspection GoVetStructTag
	type ManifestJSON struct {
		Status        int                    `json:"-"`
		Error         string                 `json:"error"`
		ErrorID       string                 `json:"error_id,omitempty"`
		ErrorCode     int                    `json:"error_code"`
		ErrorDetail   string                 `json:"error_detail,omitempty"`
		ErrorDetails  []string               `json:"error_details,omitempty"`
		customFillers []ManifestCustomFiller `json:"-"`
	}

	var w http.ResponseWriter
	var r *http.Request

	if httpCtx, ok := ctx.(HttpContext); ok {
		w, r = httpCtx.W, httpCtx.R
	} else {
		return
	}

	// Клонируем manifest если есть custom fillers
	if len(manifest.customFillers) != 0 {
		manifest = manifest.Clone()
	}

	// Вызываем custom fillers
	for i, n := 0, len(manifest.customFillers); i < n; i++ {
		manifest.customFillers[i](r, manifest)
	}

	// Отправляем JSON ответ
	jsonManifest := (*ManifestJSON)(manifest)
	sendJSON(w, jsonManifest.Status, jsonManifest)
}

// NewCommonApplicator создаёт новый CommonApplicator.
func NewCommonApplicator() *CommonApplicator {
	return new(CommonApplicator)
}

////////////////////////////////////////////////////////////////////////////////
///// CUSTOM FILLERS ///////////////////////////////////////////////////////////
////////////////////////////////////////////////////////////////////////////////

// CACF_AutoErrorDetail — ManifestCustomFiller, который извлекает ErrorDetail
// из HTTP контекста (httph.ErrorGetDetail).
//
//goland:noinspection GoSnakeCaseUsage
func CACF_AutoErrorDetail(ctx any, manifest *Manifest) {
	if r, ok := ctx.(*http.Request); ok {
		manifest.ErrorDetail = httph.ErrorGetDetail(r)
	}
}

////////////////////////////////////////////////////////////////////////////////
///// INTERNAL /////////////////////////////////////////////////////////////////
////////////////////////////////////////////////////////////////////////////////

// sendJSON отправляет JSON ответ с указанным статус кодом.
func sendJSON(w http.ResponseWriter, statusCode int, obj any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(statusCode)

	if obj != nil {
		_ = json.NewEncoder(w).Encode(obj)
	}
}
