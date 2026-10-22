package typesafe

import (
	"encoding/json"

	jsonv2 "encoding/json/v2"
)

// Request is a typed System One call: one state value and one or more questions.
type Request[State any] struct {
	Model     string
	State     State
	Questions map[string]Question
}

// MarshalJSON implements the SystemOneRequest schema.
func (r Request[State]) MarshalJSON() ([]byte, error) {
	questions, err := MarshalQuestions(r.Questions)
	if err != nil {
		return nil, err
	}
	wire := struct {
		Model     string                     `json:"model"`
		State     State                      `json:"state"`
		Questions map[string]json.RawMessage `json:"questions"`
	}{
		Model:     r.Model,
		State:     r.State,
		Questions: questions,
	}
	return jsonv2.Marshal(wire)
}
