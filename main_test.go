package main

import (
	"errors"
	"os"
	"strings"
	"testing"

	jsonv2 "encoding/json/v2"

	"github.com/hpietukhin/jevpipe/typesafe"
)

func runLive(t *testing.T, stdin string, args ...string) map[string]any {
	t.Helper()
	if os.Getenv("JEV_KEY") == "" {
		t.Skip("JEV_KEY not set; skipping live System One test")
	}
	var stdout strings.Builder
	if err := run(t.Context(), args, strings.NewReader(stdin), &stdout, os.Getenv); err != nil {
		t.Fatal(err)
	}
	var out map[string]any
	if err := jsonv2.Unmarshal([]byte(stdout.String()), &out); err != nil {
		t.Fatalf("stdout %q: %v", stdout.String(), err)
	}
	return out
}

func TestRun_live_openQuestionOnJSON_picksLeafValue(t *testing.T) {
	t.Parallel()
	out := runLive(t, `{"project_id":31935,"name":"asistent-theseus"}`, "what", "is", "the", "project", "id?")
	if out["type"] != "choice" || out["choice"] != "31935" || out["model"] == "" {
		t.Fatalf("out: %#v", out)
	}
}

func TestRun_live_questionOnText_returnsNoul(t *testing.T) {
	t.Parallel()
	yes := runLive(t, "I was charged twice for my subscription", "this mentions billing")
	no := runLive(t, "the weather is sunny today", "does this mention billing?")
	if yes["type"] != "noul" || yes["noul"].(float64) < 0.8 || no["noul"].(float64) > 0.2 {
		t.Fatalf("yes: %#v no: %#v", yes, no)
	}
}

func TestRun_noQuestion_isUsageError(t *testing.T) {
	t.Parallel()
	err := run(t.Context(), nil, strings.NewReader("{}"), &strings.Builder{}, os.Getenv)
	if !errors.Is(err, errUsage) {
		t.Fatalf("err: %v", err)
	}
}

func TestOutput_flattensFullAnswer(t *testing.T) {
	t.Parallel()
	resp := typesafe.Response{
		Model:   "jev-test",
		Answers: map[string]typesafe.Answer{"answer": typesafe.NoulAnswer{Type: typesafe.TypeNoul, Noul: 0}},
	}
	got, err := output(resp, "answer")
	if err != nil {
		t.Fatal(err)
	}
	const want = `{"model":"jev-test","type":"noul","noul":0}`
	if string(got) != want {
		t.Fatalf("got %s\nwant %s", got, want)
	}
}

func TestOutput_missingAnswer_isError(t *testing.T) {
	t.Parallel()
	if _, err := output(typesafe.Response{}, "answer"); err == nil {
		t.Fatal("expected error")
	}
}
