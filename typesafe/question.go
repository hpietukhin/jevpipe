package typesafe

import (
	"encoding/json"
	"fmt"

	jsonv2 "encoding/json/v2"
)

// QuestionType names a System One primitive.
type QuestionType string

const (
	TypeNoul   QuestionType = "noul"
	TypeChoice QuestionType = "choice"
	TypeScore  QuestionType = "score"
)

// Question is implemented only by Noul, Choice, and Score in this package.
type Question interface {
	Type() QuestionType
	marshalWire() (json.RawMessage, error)
}

// questionWire is the JSON shape TypeSafe expects for every question.
type questionWire[T any] struct {
	Type         QuestionType `json:"type"`
	Instructions string       `json:"instructions"`
	Criteria     T            `json:"criteria,omitempty"`
}

// Noul asks for P(yes) that the instructions hold about the state.
type Noul struct {
	Instructions string
	Criteria     map[string]string // optional yes/no clarification
}

func (Noul) Type() QuestionType { return TypeNoul }

func (q Noul) marshalWire() (json.RawMessage, error) {
	return jsonv2.Marshal(questionWire[map[string]string]{
		Type:         TypeNoul,
		Instructions: q.Instructions,
		Criteria:     q.Criteria,
	})
}

// Choice picks one labeled option.
type Choice struct {
	Instructions string
	Criteria     map[string]string // option -> description (empty string allowed)
}

func (Choice) Type() QuestionType { return TypeChoice }

func (q Choice) marshalWire() (json.RawMessage, error) {
	return jsonv2.Marshal(questionWire[map[string]string]{
		Type:         TypeChoice,
		Instructions: q.Instructions,
		Criteria:     q.Criteria,
	})
}

// Score rates the state along ordered levels.
type Score struct {
	Instructions string
	Criteria     []string
}

func (Score) Type() QuestionType { return TypeScore }

func (q Score) marshalWire() (json.RawMessage, error) {
	return jsonv2.Marshal(questionWire[[]string]{
		Type:         TypeScore,
		Instructions: q.Instructions,
		Criteria:     q.Criteria,
	})
}

// MarshalQuestions encodes a named set of questions for POST /v1/systemone.
func MarshalQuestions(questions map[string]Question) (map[string]json.RawMessage, error) {
	out := make(map[string]json.RawMessage, len(questions))
	for name, q := range questions {
		raw, err := q.marshalWire()
		if err != nil {
			return nil, fmt.Errorf("question %q: %w", name, err)
		}
		out[name] = raw
	}
	return out, nil
}
