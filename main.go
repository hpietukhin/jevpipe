// jev reads stdin and asks Jev (TypeSafe System One) a question about that state.
// Follows the Unix piping pattern: stdin in, JSON on stdout, errors on stderr.
package main

import (
	"cmp"
	"context"
	"fmt"
	"io"
	"os"
	"os/signal"
	"strings"
	"syscall"

	jsonv2 "encoding/json/v2"

	"github.com/havriil/jevpipe/routing"
	"github.com/havriil/jevpipe/typesafe"
)

const (
	defaultBaseURL = "https://api.typesafe.ai"
	defaultModel   = "jev-latest"
)

func main() {
	os.Exit(run())
}

func run() int {
	if len(os.Args) < 2 {
		fmt.Fprintln(os.Stderr, "usage: jev <question>")
		fmt.Fprintln(os.Stderr, "       cat data.json | jev \"what is the project id?\"")
		return 2
	}
	prompt := strings.Join(os.Args[1:], " ")

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	stdin, err := io.ReadAll(os.Stdin)
	if err != nil {
		fail("read stdin: %v", err)
		return 1
	}
	state := routing.StdinState(stdin)

	cfg, err := loadConfig()
	if err != nil {
		fail("%v", err)
		return 1
	}

	req := typesafe.Request[any]{
		Model: cfg.Model,
		State: state,
		Questions: map[string]typesafe.Question{
			"answer": routing.Question(state, prompt),
		},
	}

	client := &typesafe.Client{
		BaseURL: cfg.BaseURL,
		APIKey:  cfg.APIKey,
	}
	wire, err := client.SystemOne(ctx, req)
	if err != nil {
		fail("%v", err)
		return 1
	}

	parsed, err := wire.Parse()
	if err != nil {
		fail("parse response: %v", err)
		return 1
	}

	out, ok := typesafe.BuildCLIOutput(parsed, "answer")
	if ok {
		if err := writeJSON(os.Stdout, out); err != nil {
			fail("encode stdout: %v", err)
			return 1
		}
		return 0
	}
	fallback := struct {
		Model   string                     `json:"model"`
		Usage   map[string]any               `json:"usage,omitempty"`
		Answers map[string]typesafe.Answer `json:"answers"`
	}{
		Model:   parsed.Model,
		Usage:   parsed.Usage,
		Answers: parsed.Answers,
	}
	if err := writeJSON(os.Stdout, fallback); err != nil {
		fail("encode stdout: %v", err)
		return 1
	}
	return 0
}

func writeJSON(w io.Writer, v any) error {
	if err := jsonv2.MarshalWrite(w, v); err != nil {
		return err
	}
	_, err := w.Write([]byte{'\n'})
	return err
}

type config struct {
	APIKey  string
	BaseURL string
	Model   string
}

func loadConfig() (config, error) {
	apiKey := cmp.Or(
		strings.TrimSpace(os.Getenv("JEV_KEY")),
		strings.TrimSpace(os.Getenv("TYPESAFE_API_KEY")),
	)
	if apiKey == "" {
		return config{}, fmt.Errorf("JEV_KEY or TYPESAFE_API_KEY must be set")
	}
	model := cmp.Or(
		strings.TrimSpace(os.Getenv("JEV_MODEL")),
		strings.TrimSpace(os.Getenv("TYPESAFE_DEFAULT_MODEL")),
		defaultModel,
	)
	baseURL := strings.TrimSuffix(
		cmp.Or(strings.TrimSpace(os.Getenv("TYPESAFE_BASE_URL")), defaultBaseURL),
		"/",
	)
	return config{APIKey: apiKey, BaseURL: baseURL, Model: model}, nil
}

func fail(format string, args ...any) {
	fmt.Fprintf(os.Stderr, "jev: "+format+"\n", args...)
}
