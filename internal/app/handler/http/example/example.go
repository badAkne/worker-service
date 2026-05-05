package example

import (
	"net/http"

	"github.com/badAkne/worker-service/internal/app/entity"
	rhandler "github.com/badAkne/worker-service/internal/app/handler/http"
	"github.com/badAkne/worker-service/internal/pkg/http/binding"
	"github.com/badAkne/worker-service/internal/pkg/http/httph"
)

// handler реализует rhandler.Example.
type handler struct {
	// TODO: Добавить зависимости (modules, repositories)
	// module module.Example
}

// NewHandler создаёт новый example handler.
//
// Пример с зависимостями:
//
//	func NewHandler(mod module.Example) rhandler.Example {
//	    return &handler{module: mod}
//	}
func NewHandler() rhandler.Example {
	return &handler{}
}

// Post обрабатывает POST /v1/example.
//
// Пример запроса:
//
//	curl -X POST http://localhost:8080/v1/example \
//	  -H "Content-Type: application/json" \
//	  -d '{"message": "Hello", "count": 5, "email": "test@example.com"}'
//
// Пример ответа:
//
//	{"success": true, "data": {"message": "Hello", "count": 5, "email": "test@example.com", "processed": true}}
func (h *handler) Post(w http.ResponseWriter, r *http.Request) {
	// 1. Парсим и валидируем JSON body
	var req entity.ExampleRequest
	if err := binding.ScanAndValidateJSON(r, &req); err != nil {
		httph.ErrorApply(r, err)
		return
	}

	// 2. Бизнес-логика (здесь просто пример)
	// result, err := h.module.Process(r.Context(), req)
	// if err != nil {
	//     httph.ErrorApply(r, err)
	//     return
	// }

	// 3. Формируем и отправляем ответ
	resp := entity.ExampleResponse{
		Success: true,
		Data: entity.ExampleResponseData{
			Message:   req.Message,
			Count:     req.Count,
			Email:     req.Email,
			Processed: true,
		},
	}

	httph.SendEncoded(w, r, http.StatusOK, resp)
}

// Get обрабатывает GET /v1/example.
//
// Пример запроса:
//
//	curl "http://localhost:8080/v1/example?message=Hello&limit=10"
//
// Пример ответа:
//
//	{"success": true, "data": {"message": "Hello", "count": 10, "processed": true}}
func (h *handler) Get(w http.ResponseWriter, r *http.Request) {
	// 1. Парсим и валидируем query параметры
	var req entity.ExampleQueryRequest
	if err := binding.ScanAndValidateQuery(r, &req); err != nil {
		httph.ErrorApply(r, err)
		return
	}

	// 2. Бизнес-логика
	// ...

	// 3. Формируем и отправляем ответ
	resp := entity.ExampleResponse{
		Success: true,
		Data: entity.ExampleResponseData{
			Message:   req.Message,
			Count:     req.Limit,
			Processed: true,
		},
	}

	httph.SendEncoded(w, r, http.StatusOK, resp)
}
