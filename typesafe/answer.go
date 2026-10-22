package typesafe

import (
	"encoding/json"
	"fmt"

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

// DecodeAnswer unmarshals one answer blob into a concrete Answer type.
func DecodeAnswer[T Answer](raw json.RawMessage) (T, error) {
	var answer T
	if err := jsonv2.Unmarshal(raw, &answer); err != nil {
		return answer, err
	}
	return answer, nil
}

// DecodeAnswers parses every answer in a System One response by declared type.
func DecodeAnswers(wire map[string]json.RawMessage) (map[string]Answer, error) {
	out := make(map[string]Answer, len(wire))
	for name, raw := range wire {
		typ, err := peekAnswerType(raw)
		if err != nil {
			return nil, fmt.Errorf("answer %q: %w", name, err)
		}
		switch typ {
		case TypeNoul:
			a, err := DecodeAnswer[NoulAnswer](raw)
			if err != nil {
				return nil, fmt.Errorf("answer %q: %w", name, err)
			}
			out[name] = a
		case TypeChoice:
			a, err := DecodeAnswer[ChoiceAnswer](raw)
			if err != nil {
				return nil, fmt.Errorf("answer %q: %w", name, err)
			}
			out[name] = a
		case TypeScore:
			a, err := DecodeAnswer[ScoreAnswer](raw)
			if err != nil {
				return nil, fmt.Errorf("answer %q: %w", name, err)
			}
			out[name] = a
		default:
			return nil, fmt.Errorf("answer %q: unknown type %q", name, typ)
		}
	}
	return out, nil
}

func peekAnswerType(raw json.RawMessage) (QuestionType, error) {
	var head struct {
		Type QuestionType `json:"type"`
	}
	if err := jsonv2.Unmarshal(raw, &head); err != nil {
		return "", err
	}
	return head.Type, nil
}

// CLIOutput is the compact JSON object written by the jev CLI.
type CLIOutput struct {
	Model         string             `json:"model"`
	Usage         map[string]any     `json:"usage,omitempty"`
	Type          QuestionType       `json:"type,omitempty,omitzero"`
	Noul          float64            `json:"noul,omitzero"`
	Choice        string             `json:"choice,omitempty"`
	Score         float64            `json:"score,omitzero"`
	Confidence    float64            `json:"confidence,omitzero"`
	Probabilities map[string]float64 `json:"probabilities,omitempty"`
}

// BuildCLIOutput projects a parsed response into CLI JSON for one named answer.
func BuildCLIOutput(parsed ParsedResponse, primary string) (CLIOutput, bool) {
	out := CLIOutput{
		Model: parsed.Model,
		Usage: parsed.Usage,
	}
	answer, ok := parsed.Answers[primary]
	if !ok {
		return out, false
	}
	out.Type = answer.AnswerType()
	switch a := answer.(type) {
	case NoulAnswer:
		out.Noul = a.Noul
	case ChoiceAnswer:
		out.Choice = a.Choice
		out.Confidence = a.Confidence
		out.Probabilities = a.Probabilities
	case ScoreAnswer:
		out.Score = a.Score
		out.Confidence = a.Confidence
	}
	return out, true
}
