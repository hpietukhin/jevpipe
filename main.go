// jev reads stdin and asks Jev (TypeSafe System One) a question about that state.
// Follows the Unix piping pattern: stdin in, JSON on stdout, errors on stderr.
package main

import (
	"cmp"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"encoding/json/jsontext"
	jsonv2 "encoding/json/v2"

	"github.com/hpietukhin/jevpipe/routing"
	"github.com/hpietukhin/jevpipe/typesafe"
)

const (
	defaultModel = "jev-latest"
	answerName   = "answer"
	usage        = "usage: jev <question>\n       cat data.json | jev \"what is the project id?\""
)

var errUsage = errors.New("usage")

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	err := run(ctx, os.Args[1:], os.Stdin, os.Stdout, os.Getenv)
	stop()
	switch {
	case err == nil:
	case errors.Is(err, errUsage):
		fmt.Fprintln(os.Stderr, usage)
		os.Exit(2)
	default:
		fmt.Fprintf(os.Stderr, "jev: %v\n", err)
		os.Exit(1)
	}
}

func run(ctx context.Context, args []string, stdin io.Reader, stdout io.Writer, getenv func(string) string) error {
	if len(args) == 0 {
		return errUsage
	}
	prompt := strings.Join(args, " ")

	data, err := io.ReadAll(stdin)
	if err != nil {
		return fmt.Errorf("read stdin: %w", err)
	}
	state := routing.StdinState(data)

	apiKey := strings.TrimSpace(getenv("JEV_KEY"))
	if apiKey == "" {
		return errors.New("JEV_KEY must be set")
	}
	client := &typesafe.Client{BaseURL: strings.TrimSpace(getenv("JEV_BASE_URL")), APIKey: apiKey}
	resp, err := client.SystemOne(ctx, typesafe.Request{
		Model:     cmp.Or(strings.TrimSpace(getenv("JEV_MODEL")), defaultModel),
		State:     state,
		Questions: map[string]typesafe.Question{answerName: routing.Question(state, prompt)},
	})
	if err != nil {
		return err
	}

	out, err := output(resp, answerName)
	if err != nil {
		return err
	}
	if _, err := stdout.Write(append(out, '\n')); err != nil {
		return fmt.Errorf("write stdout: %w", err)
	}
	return nil
}

// output flattens model, usage, and every field of the named answer into one JSON object.
func output(resp typesafe.Response, name string) ([]byte, error) {
	answer, ok := resp.Answers[name]
	if !ok {
		return nil, fmt.Errorf("response has no %q answer", name)
	}
	fields, err := jsonv2.Marshal(answer)
	if err != nil {
		return nil, fmt.Errorf("encode answer: %w", err)
	}
	return jsonv2.Marshal(struct {
		Model  string         `json:"model"`
		Usage  map[string]any `json:"usage,omitempty"`
		Answer jsontext.Value `json:",embed"`
	}{resp.Model, resp.Usage, fields})
}
