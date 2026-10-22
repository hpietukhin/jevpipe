package typesafe

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"

	jsonv2 "encoding/json/v2"
)

const (
	SystemOnePath  = "/v1/systemone"
	maxErrorBody   = 200
	defaultTimeout = 120 * time.Second
)

// Client calls the TypeSafe HTTP API.
type Client struct {
	BaseURL    string
	APIKey     string
	HTTPClient *http.Client
}

var defaultHTTPClient = sync.OnceValue(func() *http.Client {
	return &http.Client{Timeout: defaultTimeout}
})

func (c *Client) httpClient() *http.Client {
	if c.HTTPClient != nil {
		return c.HTTPClient
	}
	return defaultHTTPClient()
}

// Response is the decoded System One payload.
type Response struct {
	Model   string                     `json:"model"`
	Usage   map[string]any             `json:"usage"`
	Answers map[string]json.RawMessage `json:"answers"`
}

// SystemOne sends a typed request and returns the raw response wire.
func (c *Client) SystemOne[State any](ctx context.Context, req Request[State]) (Response, error) {
	payload, err := jsonv2.Marshal(req)
	if err != nil {
		return Response{}, fmt.Errorf("marshal request: %w", err)
	}

	url := c.BaseURL + SystemOnePath
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(payload))
	if err != nil {
		return Response{}, fmt.Errorf("build request: %w", err)
	}
	httpReq.Header.Set("Authorization", "Bearer "+c.APIKey)
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Accept", "application/json")

	resp, err := c.httpClient().Do(httpReq)
	if err != nil {
		return Response{}, fmt.Errorf("http: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return Response{}, fmt.Errorf("read response: %w", err)
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return Response{}, fmt.Errorf("api %s: %s", resp.Status, truncateBody(body))
	}

	var parsed Response
	if err := jsonv2.Unmarshal(body, &parsed); err != nil {
		return Response{}, fmt.Errorf("decode response: %w", err)
	}
	return parsed, nil
}

func truncateBody(body []byte) string {
	s := string(body)
	if len(s) <= maxErrorBody {
		return s
	}
	return s[:maxErrorBody] + "…"
}

// ParsedResponse pairs wire metadata with typed answers.
type ParsedResponse struct {
	Model   string
	Usage   map[string]any
	Answers map[string]Answer
}

// Parse decodes every answer in a Response using primitive-specific types.
func (r Response) Parse() (ParsedResponse, error) {
	answers, err := DecodeAnswers(r.Answers)
	if err != nil {
		return ParsedResponse{}, fmt.Errorf("parse answers: %w", err)
	}
	return ParsedResponse{
		Model:   r.Model,
		Usage:   r.Usage,
		Answers: answers,
	}, nil
}
