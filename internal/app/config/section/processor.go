package section

type (
	// Processor — конфигурация процессоров.
	Processor struct {
		WebServer ProcessorWebServer `split_words:"true"`
	}

	// ProcessorWebServer — конфигурация HTTP веб-сервера.
	ProcessorWebServer struct {
		ListenPort uint32 `split_words:"true" default:"8080"`
	}
)
