package typesafe

import (
	"fmt"

	"encoding/json/jsontext"
	jsonv2 "encoding/json/v2"
)

// Answer is implemented by each primitive returned under Response.Answers.
type Answer interface {
	AnswerType() QuestionType
}

// NoulAnswer is P(yes) for a Noul question.
type NoulAnswer struct {
	Type QuestionType `json:"type"`
	Noul float64      `json:"noul"`
}

func (a NoulAnswer) AnswerType() QuestionType { return TypeNoul }

// ChoiceAnswer is the selected option for a Choice question.
type ChoiceAnswer struct {
	Type          QuestionType       `json:"type"`
	Choice        string             `json:"choice"`
	Confidence    float64            `json:"confidence"`
	Probabilities map[string]float64 `json:"probabilities"`
}

func (a ChoiceAnswer) AnswerType() QuestionType { return TypeChoice }

// ScoreAnswer is the position along Score levels.
type ScoreAnswer struct {
	Type          QuestionType       `json:"type"`
	Score         float64            `json:"score"`
	Confidence    float64            `json:"confidence"`
	Legend        map[string]string  `json:"legend"`
	Probabilities map[string]float64 `json:"probabilities"`
}

func (a ScoreAnswer) AnswerType() QuestionType { return TypeScore }

var answerDecoders = map[QuestionType]func(jsontext.Value) (Answer, error){
	TypeNoul:   decodeAs[NoulAnswer],
	TypeChoice: decodeAs[ChoiceAnswer],
	TypeScore:  decodeAs[ScoreAnswer],
}

func decodeAs[T Answer](raw jsontext.Value) (Answer, error) {
	var answer T
	err := jsonv2.Unmarshal(raw, &answer)
	return answer, err
}

func decodeAnswer(raw jsontext.Value) (Answer, error) {
	var head struct {
		Type QuestionType `json:"type"`
	}
	if err := jsonv2.Unmarshal(raw, &head); err != nil {
		return nil, err
	}
	decode, ok := answerDecoders[head.Type]
	if !ok {
		return nil, fmt.Errorf("unknown type %q", head.Type)
	}
	return decode(raw)
}
