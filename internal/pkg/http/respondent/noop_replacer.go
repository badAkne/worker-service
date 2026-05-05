package respondent

// noOpReplacer — реализация Replacer, которая ничего не делает.
// Просто возвращает переданную ошибку без изменений.
type noOpReplacer struct{}

// Replace возвращает переданную ошибку без изменений.
func (*noOpReplacer) Replace(err error) error {
	return err
}

func newNoOpReplacer() *noOpReplacer {
	return new(noOpReplacer)
}
