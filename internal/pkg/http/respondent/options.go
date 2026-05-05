package respondent

// Option — функция для настройки Respondent.
type Option func(m *respondent)

// WithReplacer устанавливает кастомный Replacer вместо дефолтного (no-op).
func WithReplacer(replacer Replacer) Option {
	return func(r *respondent) {
		if replacer != nil {
			r.fromOptions.replacer = replacer
		}
	}
}

// WithApplicator устанавливает кастомный Applicator вместо дефолтного.
func WithApplicator(applicator Applicator) Option {
	return func(r *respondent) {
		if applicator != nil {
			r.fromOptions.applicator = applicator
		}
	}
}
