package respondent

// Expander — интерфейс для преобразования ошибки в Manifest.
// Реализация должна распознать тип ошибки и вернуть соответствующий Manifest.
type Expander interface {
	Expand(err error) *Manifest
}

// Replacer — интерфейс для замены одной ошибки на другую.
type Replacer interface {
	Replace(err error) error
}

// Applicator — интерфейс для преобразования Manifest в HTTP ответ.
type Applicator interface {
	Apply(ctx any, manifest *Manifest)
}

// Manifest — структурированное представление ошибки для HTTP ответа.
type Manifest struct {
	Status    int    // HTTP статус код
	Error     string // Сообщение об ошибке
	ErrorID   string // Уникальный ID ошибки (опционально)
	ErrorCode int    // Код ошибки (например, 40401)

	ErrorDetail  string   // Дополнительные детали (одна строка)
	ErrorDetails []string // Дополнительные детали (массив)

	customFillers []ManifestCustomFiller
}

// ManifestCustomFiller — функция для кастомного заполнения Manifest.
type ManifestCustomFiller = func(ctx any, manifest *Manifest)

// Clone возвращает копию Manifest.
func (m *Manifest) Clone() *Manifest {
	if m != nil {
		m2 := *m
		return &m2
	}
	return nil
}
