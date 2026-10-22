package typesafe

import (
	jsonv2 "encoding/json/v2"
)

// Request is a System One call: one state value and one or more questions.
type Request struct {
	Model     string
	State     any
	Questions map[string]Question
}

// MarshalJSON implements the SystemOneRequest schema.
func (r Request) MarshalJSON() ([]byte, error) {
	questions := make(map[string]any, len(r.Questions))
	for name, q := range r.Questions {
		questions[name] = q.wire()
	}
	return jsonv2.Marshal(struct {
		Model     string         `json:"model"`
		State     any            `json:"state"`
		Questions map[string]any `json:"questions"`
	}{r.Model, r.State, questions})
}
