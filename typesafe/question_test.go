package typesafe

import (
	"encoding/json"
	"strings"
	"testing"

	jsonv2 "encoding/json/v2"
)

func TestQuestionWire_roundTripsType(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name     string
		question Question
		wantType QuestionType
	}{
		{
			name:     "noul",
			question: Noul{Instructions: "Is this about billing?"},
			wantType: TypeNoul,
		},
		{
			name: "choice",
			question: Choice{
				Instructions: "Which id?",
				Criteria:     map[string]string{"42": "project_id", "none": "escape"},
			},
			wantType: TypeChoice,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			raw, err := jsonv2.Marshal(tc.question.wire())
			if err != nil {
				t.Fatal(err)
			}
			var head struct {
				Type QuestionType `json:"type"`
			}
			if err := jsonv2.Unmarshal(raw, &head); err != nil {
				t.Fatal(err)
			}
			if head.Type != tc.wantType {
				t.Fatalf("type: %v", head.Type)
			}
		})
	}
}

func TestRequestMarshalJSON_includesQuestionType(t *testing.T) {
	t.Parallel()
	req := Request{
		Model: "jev-latest",
		State: map[string]int{"project_id": 42},
		Questions: map[string]Question{
			"answer": Noul{Instructions: "Is project_id set?"},
		},
	}
	data, err := jsonv2.Marshal(req)
	if err != nil {
		t.Fatal(err)
	}
	if !json.Valid(data) {
		t.Fatalf("invalid json: %s", data)
	}
	if !strings.Contains(string(data), `"type":"noul"`) {
		t.Fatalf("missing noul question in body: %s", data)
	}
}

func TestDecodeResponse_parsesNoulAndChoice(t *testing.T) {
	t.Parallel()
	resp, err := decodeResponse([]byte(`{"model":"m","answers":{
		"billing": {"type":"noul","noul":0.9},
		"tone": {"type":"choice","choice":"calm","confidence":0.8,"probabilities":{"calm":0.8}}
	}}`))
	if err != nil {
		t.Fatal(err)
	}
	answers := resp.Answers
	noul, ok := answers["billing"].(NoulAnswer)
	if !ok || noul.Noul != 0.9 {
		t.Fatalf("noul: %#v", answers["billing"])
	}
	choice, ok := answers["tone"].(ChoiceAnswer)
	if !ok || choice.Choice != "calm" {
		t.Fatalf("choice: %#v", answers["tone"])
	}
}
