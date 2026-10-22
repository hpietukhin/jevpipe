package routing

import (
	"encoding/json"
	"testing"

	"github.com/havriil/jevpipe/typesafe"
)

func TestStdinState_largeIntegerStaysExact(t *testing.T) {
	t.Parallel()
	state := StdinState([]byte(`{"id":9007199254740993}`))
	m, ok := state.(map[string]any)
	if !ok {
		t.Fatalf("state: %#v", state)
	}
	id, ok := m["id"].(json.Number)
	if !ok {
		t.Fatalf("id type: %T", m["id"])
	}
	if id.String() != "9007199254740993" {
		t.Fatalf("id: %s", id)
	}
}

func TestStdinState_invalidJSON_becomesString(t *testing.T) {
	t.Parallel()
	if got := StdinState([]byte(`not-json`)); got != "not-json" {
		t.Fatalf("got %#v", got)
	}
}

func TestScalarCandidates_recordsJSONPaths(t *testing.T) {
	t.Parallel()
	state := map[string]any{
		"project_id": 31935,
		"name":       "theseus",
	}
	byValue := map[string]string{}
	for _, c := range scalarCandidates(state) {
		byValue[c.Value] = c.Path
	}
	if byValue["31935"] != "project_id" {
		t.Fatalf("paths: %#v", byValue)
	}
	if byValue["theseus"] != "name" {
		t.Fatalf("paths: %#v", byValue)
	}
}

func TestQuestion_openPromptOnJSON_usesChoiceWithPaths(t *testing.T) {
	t.Parallel()
	state := map[string]any{"project_id": 42}
	q := Question(state, "what is the project id?")
	ch, ok := q.(typesafe.Choice)
	if !ok {
		t.Fatalf("type: %T", q)
	}
	if ch.Criteria["42"] != "project_id" {
		t.Fatalf("criteria: %#v", ch.Criteria)
	}
}

func TestQuestion_yesNoPrompt_usesNoul(t *testing.T) {
	t.Parallel()
	state := map[string]any{"project_id": 42}
	q := Question(state, "is project_id set?")
	if q.Type() != typesafe.TypeNoul {
		t.Fatalf("type: %v", q.Type())
	}
}
