package routing

import (
	"encoding/json"
	"fmt"
	"maps"
	"slices"
	"testing"

	"github.com/hpietukhin/jevpipe/typesafe"
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

func TestStdinState_textStartingWithJSON_staysString(t *testing.T) {
	t.Parallel()
	if got := StdinState([]byte(`3 apples`)); got != "3 apples" {
		t.Fatalf("got %#v", got)
	}
}

func TestQuestion_openPromptOnScalarState_usesNoul(t *testing.T) {
	t.Parallel()
	for _, stdin := range []string{"I was charged twice", "42", `"quoted"`, ""} {
		if q := Question(StdinState([]byte(stdin)), "what is this about?"); q.Type() != typesafe.TypeNoul {
			t.Fatalf("stdin %q: got %v", stdin, q.Type())
		}
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

func TestQuestion_candidatesPastCap_areDeterministic(t *testing.T) {
	t.Parallel()
	state := map[string]any{}
	for i := range 3 * maxCandidates {
		state[fmt.Sprintf("k%03d", i)] = json.Number(fmt.Sprint(i))
	}
	first := Question(state, "which key?").(typesafe.Choice).Criteria
	for range 20 {
		got := Question(state, "which key?").(typesafe.Choice).Criteria
		if !maps.Equal(got, first) {
			t.Fatalf("criteria changed between runs")
		}
	}
}

func TestQuestion_noneValue_keepsSentinelSeparate(t *testing.T) {
	t.Parallel()
	q := Question(map[string]any{"status": "none"}, "what status?")
	criteria := q.(typesafe.Choice).Criteria
	if criteria["none"] != "status" {
		t.Fatalf("data candidate: %#v", criteria)
	}
	if !slices.Contains(slices.Collect(maps.Values(criteria)), noneDescription) {
		t.Fatalf("sentinel missing: %#v", criteria)
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
