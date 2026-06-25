package fixer

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"github.com/badAkne/worker-service/internal/app/config/section"
	"github.com/badAkne/worker-service/internal/app/entity"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
)

type Client struct {
	httpClient *http.Client
	apiKey     string
	baseURL    string
}

type FixerResponse struct {
	Success   bool               `json:"success"`
	Timestamp int64              `json:"timestamp"`
	Base      string             `json:"base"`
	Date      string             `json:"date"`
	Rates     map[string]float64 `json:"rates"`
	Error     *FixerError        `json:"error,omitempty"`
}

type FixerError struct {
	Code int    `json:"code"`
	Type string `json:"type"`
	Info string `json:"info"`
}

func NewClient(cfg section.ClientFixer) *Client {
	httpClient := &http.Client{
		Transport: otelhttp.NewTransport(http.DefaultTransport),
		Timeout:   10 * time.Second,
	}

	return &Client{
		httpClient: httpClient,
		baseURL:    cfg.BaseURL,
		apiKey:     cfg.ApiKey,
	}
}

func (c *Client) GetRates(ctx context.Context, base string) (map[string]float64, error) {
	args := make(url.Values)
	args.Set("access_key", c.apiKey)
	args.Set("base", base)

	requestURL := fmt.Sprintf("%s/latest?%s", c.baseURL, args.Encode())

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, requestURL, http.NoBody)
	if err != nil {
		return nil, fmt.Errorf("unable to make a request: %w", err)
	}

	res, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("unable to do request to fixer api: %w", err)
	}
	defer res.Body.Close()

	if res.StatusCode == http.StatusServiceUnavailable {
		return nil, entity.ErrFixerUnavailable
	}

	var fixerRes FixerResponse

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(body, &fixerRes)
	if err != nil {
		return nil, entity.ErrFixerInvalidResponse
	}

	if !fixerRes.Success {
		var fixErr FixerError
		_ = json.Unmarshal(body, &fixErr)
		return nil, c.mapFixerError(&fixErr)
	}

	return fixerRes.Rates, nil
}

// mapFixerError преобразует ошибку Fixer API в типизированную ошибку.
func (c *Client) mapFixerError(fixerErr *FixerError) error {
	if fixerErr == nil {
		return entity.ErrFixerInvalidResponse
	}

	switch fixerErr.Code {
	case 101: // invalid_access_key
		return entity.ErrFixerInvalidApiKey
	case 104, 105: // rate limit
		return entity.ErrFixerRateLimitExceeded
	default:
		return fmt.Errorf("%w: [%d] %s - %s", entity.ErrFixerInvalidResponse,
			fixerErr.Code, fixerErr.Type, fixerErr.Info)
	}
}
