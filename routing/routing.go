// Package routing turns stdin state and a user prompt into a TypeSafe Question.
package routing

import (
	"bytes"
	"encoding/json"
	"fmt"
	"regexp"
	"strings"

	"github.com/havriil/jevpipe/typesafe"
)

const maxCandidates = 200

var yesNoPrompt = regexp.MustCompile(`(?i)^(is|are|was|were|does|do|did|has|have|had|can|could|should|would|will)\b`)

// StdinState is parsed stdin content ready for System One state.
// JSON uses json.Number for integers so large IDs stay exact; non-JSON input is returned as a string.
func StdinState(data []byte) any {
	trimmed := bytes.TrimSpace(data)
	if len(trimmed) == 0 {
		return ""
	}
	dec := json.NewDecoder(bytes.NewReader(trimmed))
	dec.UseNumber()
	var parsed any
	if err := dec.Decode(&parsed); err == nil {
		return parsed
	}
	return string(trimmed)
}

// Question picks Noul vs Choice from stdin state and the user's words.
func Question(state any, prompt string) typesafe.Question {
	candidates := scalarCandidates(state)
	if len(candidates) > 0 && !yesNoPrompt.MatchString(strings.TrimSpace(prompt)) {
		criteria := map[string]string{
			"none": "None of the candidate values answers the question.",
		}
		for _, c := range candidates {
			if prev, ok := criteria[c.Value]; ok && prev != "" {
				criteria[c.Value] = prev + "; " + c.Path
			} else {
				criteria[c.Value] = c.Path
			}
		}
		return typesafe.Choice{
			Instructions: prompt,
			Criteria:     criteria,
		}
	}
	return typesafe.Noul{Instructions: prompt}
}

// ScalarCandidate is a leaf value in stdin state and its JSON path.
type ScalarCandidate struct {
	Path  string
	Value string
}

func scalarCandidates(state any) []ScalarCandidate {
	seen := make(map[string]struct{})
	var out []ScalarCandidate
	walk(state, "", seen, &out)
	return out
}

func walk(v any, path string, seen map[string]struct{}, out *[]ScalarCandidate) {
	if len(*out) >= maxCandidates {
		return
	}
	switch val := v.(type) {
	case map[string]any:
		for k, child := range val {
			childPath := k
			if path != "" {
				childPath = path + "." + k
			}
			walk(child, childPath, seen, out)
		}
	case []any:
		for i, child := range val {
			childPath := fmt.Sprintf("%s[%d]", path, i)
			if path == "" {
				childPath = fmt.Sprintf("[%d]", i)
			}
			walk(child, childPath, seen, out)
		}
	case json.Number:
		addCandidate(path, val.String(), seen, out)
	default:
		s := scalarString(val)
		if s == "" {
			return
		}
		addCandidate(path, s, seen, out)
	}
}

func addCandidate(path, value string, seen map[string]struct{}, out *[]ScalarCandidate) {
	key := path + "\x00" + value
	if _, ok := seen[key]; ok {
		return
	}
	seen[key] = struct{}{}
	if path == "" {
		path = "."
	}
	*out = append(*out, ScalarCandidate{Path: path, Value: value})
}

func scalarString(v any) string {
	switch t := v.(type) {
	case nil:
		return ""
	case string:
		return strings.TrimSpace(t)
	case float64:
		if t == float64(int64(t)) {
			return fmt.Sprintf("%d", int64(t))
		}
		return fmt.Sprintf("%g", t)
	case bool:
		if t {
			return "true"
		}
		return "false"
	default:
		return strings.TrimSpace(fmt.Sprint(v))
	}
}
