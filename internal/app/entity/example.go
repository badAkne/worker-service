package entity

// ExampleRequest — пример request структуры для JSON body.
// Теги:
//   - json:"field" — имя поля в JSON
//   - binding:"required" — обязательное поле
//   - binding:"email" — валидация email
//   - binding:"min=1,max=100" — ограничения длины
type ExampleRequest struct {
	// Message — обязательное текстовое поле
	Message string `json:"message" binding:"required,min=1,max=1000"`

	// Count — опциональное число (по умолчанию 0)
	Count int `json:"count" binding:"min=0,max=100"`

	// Email — опциональный email с валидацией
	Email string `json:"email,omitempty" binding:"omitempty,email"`
}

// ExampleQueryRequest — пример request структуры для query параметров.
// Теги:
//   - form:"field" — имя параметра в URL (?field=value)
//   - binding:"required" — обязательный параметр
type ExampleQueryRequest struct {
	// Message — обязательный query параметр
	Message string `form:"message" binding:"required"`

	// Limit — опциональный лимит
	Limit int `form:"limit" binding:"min=0,max=100"`
}

// ExampleResponse — пример response структуры.
type ExampleResponse struct {
	// Success — флаг успешности
	Success bool `json:"success"`

	// Data — данные ответа
	Data ExampleResponseData `json:"data"`
}

// ExampleResponseData — данные в ответе.
type ExampleResponseData struct {
	Message   string `json:"message"`
	Count     int    `json:"count,omitempty"`
	Email     string `json:"email,omitempty"`
	Processed bool   `json:"processed"`
}
