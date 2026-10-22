package typesafe

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
	wire() any
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

func (q Noul) wire() any {
	return questionWire[map[string]string]{TypeNoul, q.Instructions, q.Criteria}
}

// Choice picks one labeled option.
type Choice struct {
	Instructions string
	Criteria     map[string]string // option -> description (empty string allowed)
}

func (Choice) Type() QuestionType { return TypeChoice }

func (q Choice) wire() any {
	return questionWire[map[string]string]{TypeChoice, q.Instructions, q.Criteria}
}

// Score rates the state along ordered levels.
type Score struct {
	Instructions string
	Criteria     []string
}

func (Score) Type() QuestionType { return TypeScore }

func (q Score) wire() any {
	return questionWire[[]string]{TypeScore, q.Instructions, q.Criteria}
}
