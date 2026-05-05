package section

type (
	// Monitor — конфигурация мониторинга и логирования.
	Monitor struct {
		LogLevel    string `default:"debug" split_words:"true"`
		Environment string `default:"dev"`
	}
)
