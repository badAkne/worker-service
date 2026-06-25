package section

import "time"

type (
	Client struct {
		Fixer ClientFixer
	}

	ClientFixer struct {
		ApiKey   string        `required:"true" split_words:"true"`
		BaseURL  string        `default:"http://data.fixer.io/api"`
		CacheTTL time.Duration `default:"30m"`
	}
)
