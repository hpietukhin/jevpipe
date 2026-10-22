package typesafe

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	jsonv2 "encoding/json/v2"
)

func TestClientSystemOne_returnsParsedNoulAnswer(t *testing.T) {
	t.Parallel()

	var captured struct {
		Authorization string
		Body          map[string]any
	}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		captured.Authorization = r.Header.Get("Authorization")
		body, err := io.ReadAll(r.Body)
		if err != nil {
			t.Errorf("read body: %v", err)
		}
		if err := jsonv2.Unmarshal(body, &captured.Body); err != nil {
			t.Errorf("decode request: %v", err)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"model":"jev-test","usage":{"input_tokens":1},"answers":{"answer":{"type":"noul","noul":0.75}}}`))
	}))
	defer server.Close()

	client := &Client{
		BaseURL:    server.URL,
		APIKey:     "test-key",
		HTTPClient: server.Client(),
	}
	req := Request[map[string]string]{
		Model: "jev-latest",
		State: map[string]string{"ticket": "duplicate charge"},
		Questions: map[string]Question{
			"answer": Noul{Instructions: "Does the ticket mention billing?"},
		},
	}

	wire, err := client.SystemOne(t.Context(), req)
	if err != nil {
		t.Fatal(err)
	}
	parsed, err := wire.Parse()
	if err != nil {
		t.Fatal(err)
	}
	noul, ok := parsed.Answers["answer"].(NoulAnswer)
	if !ok || noul.Noul != 0.75 {
		t.Fatalf("answer: %#v", parsed.Answers["answer"])
	}
	if captured.Authorization != "Bearer test-key" {
		t.Fatalf("authorization: %q", captured.Authorization)
	}
	if captured.Body["model"] != "jev-latest" {
		t.Fatalf("request model: %v", captured.Body["model"])
	}
}

func TestClientSystemOneHTTPError_truncatesBodyInError(t *testing.T) {
	t.Parallel()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		http.Error(w, strings.Repeat("e", maxErrorBody+100), http.StatusBadRequest)
	}))
	defer server.Close()

	client := &Client{BaseURL: server.URL, APIKey: "k", HTTPClient: server.Client()}
	_, err := client.SystemOne(t.Context(), Request[string]{
		Model: "m",
		State: "",
		Questions: map[string]Question{
			"a": Noul{Instructions: "q"},
		},
	})
	if err == nil {
		t.Fatal("expected error")
	}
	const prefix = "api 400 Bad Request: "
	if len(err.Error()) > len(prefix)+maxErrorBody+len("…")+5 {
		t.Fatalf("error too long (%d chars): %s", len(err.Error()), err.Error()[:80])
	}
}
