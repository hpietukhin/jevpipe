package typesafe

import (
	"bytes"
	"cmp"
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"

	"encoding/json/jsontext"
	jsonv2 "encoding/json/v2"
)

const (
	DefaultBaseURL = "https://api.typesafe.ai"
	systemOnePath  = "/v1/systemone"
	maxErrorBody   = 200
	defaultTimeout = 120 * time.Second
)

// Client calls the TypeSafe HTTP API.
type Client struct {
	BaseURL    string // DefaultBaseURL when empty
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

// Response is a System One reply with every answer decoded to its primitive type.
type Response struct {
	Model   string
	Usage   map[string]any
	Answers map[string]Answer
}

// SystemOne sends req to POST /v1/systemone and decodes the answers.
func (c *Client) SystemOne(ctx context.Context, req Request) (Response, error) {
	payload, err := jsonv2.Marshal(req)
	if err != nil {
		return Response{}, fmt.Errorf("marshal request: %w", err)
	}

	url := strings.TrimSuffix(cmp.Or(c.BaseURL, DefaultBaseURL), "/") + systemOnePath
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
	return decodeResponse(body)
}

func decodeResponse(body []byte) (Response, error) {
	var wire struct {
		Model   string                    `json:"model"`
		Usage   map[string]any            `json:"usage"`
		Answers map[string]jsontext.Value `json:"answers"`
	}
	if err := jsonv2.Unmarshal(body, &wire); err != nil {
		return Response{}, fmt.Errorf("decode response: %w", err)
	}
	answers := make(map[string]Answer, len(wire.Answers))
	for name, raw := range wire.Answers {
		answer, err := decodeAnswer(raw)
		if err != nil {
			return Response{}, fmt.Errorf("answer %q: %w", name, err)
		}
		answers[name] = answer
	}
	return Response{Model: wire.Model, Usage: wire.Usage, Answers: answers}, nil
}

func truncateBody(body []byte) string {
	s := string(body)
	if len(s) <= maxErrorBody {
		return s
	}
	return s[:maxErrorBody] + "…"
}
