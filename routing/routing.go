// Package routing turns stdin state and a user prompt into a TypeSafe Question.
package routing

import (
	"bytes"
	"encoding/json"
	"fmt"
	"maps"
	"regexp"
	"slices"
	"strings"

	"github.com/hpietukhin/jevpipe/typesafe"
)

const (
	maxCandidates   = 200
	noneDescription = "None of the candidate values answers the question."
)

var yesNoPrompt = regexp.MustCompile(`(?i)^(is|are|was|were|does|do|did|has|have|had|can|could|should|would|will)\b`)

// StdinState is parsed stdin content ready for System One state.
// JSON uses json.Number for integers so large IDs stay exact; non-JSON input is returned as a string.
func StdinState(data []byte) any {
	trimmed := bytes.TrimSpace(data)
	if !json.Valid(trimmed) {
		return string(trimmed)
	}
	dec := json.NewDecoder(bytes.NewReader(trimmed))
	dec.UseNumber()
	var parsed any
	if err := dec.Decode(&parsed); err != nil {
		return string(trimmed)
	}
	return parsed
}

// Question asks Choice over leaf values when state is a JSON object or array and
// the prompt is open; everything else is a Noul.
func Question(state any, prompt string) typesafe.Question {
	candidates := scalarCandidates(state)
	if len(candidates) == 0 || yesNoPrompt.MatchString(strings.TrimSpace(prompt)) {
		return typesafe.Noul{Instructions: prompt}
	}
	criteria := map[string]string{
		noneKey(candidates): noneDescription,
	}
	for _, c := range candidates {
		if prev, ok := criteria[c.Value]; ok {
			criteria[c.Value] = prev + "; " + c.Path
		} else {
			criteria[c.Value] = c.Path
		}
	}
	return typesafe.Choice{Instructions: prompt, Criteria: criteria}
}

func noneKey(candidates []scalarCandidate) string {
	key := "none"
	for slices.ContainsFunc(candidates, func(c scalarCandidate) bool { return c.Value == key }) {
		key = "_" + key
	}
	return key
}

// scalarCandidate is a leaf value in stdin state and its JSON path.
type scalarCandidate struct {
	Path  string
	Value string
}

func scalarCandidates(state any) []scalarCandidate {
	switch state.(type) {
	case map[string]any, []any:
	default:
		return nil
	}
	var out []scalarCandidate
	walk(state, "", &out)
	return out
}

func walk(v any, path string, out *[]scalarCandidate) {
	if len(*out) >= maxCandidates {
		return
	}
	switch val := v.(type) {
	case map[string]any:
		for _, k := range slices.Sorted(maps.Keys(val)) {
			childPath := k
			if path != "" {
				childPath = path + "." + k
			}
			walk(val[k], childPath, out)
		}
	case []any:
		for i, child := range val {
			walk(child, fmt.Sprintf("%s[%d]", path, i), out)
		}
	case nil:
	default:
		value := strings.TrimSpace(fmt.Sprint(val))
		if value == "" {
			return
		}
		*out = append(*out, scalarCandidate{Path: path, Value: value})
	}
}
