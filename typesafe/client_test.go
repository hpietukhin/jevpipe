package typesafe

import (
	"os"
	"strings"
	"testing"
)

func liveClient(t *testing.T) *Client {
	t.Helper()
	key := os.Getenv("JEV_KEY")
	if key == "" {
		t.Skip("JEV_KEY not set; skipping live System One test")
	}
	return &Client{BaseURL: os.Getenv("JEV_BASE_URL"), APIKey: key}
}

func TestClientSystemOne_live_decodesEveryPrimitive(t *testing.T) {
	t.Parallel()
	client := liveClient(t)
	resp, err := client.SystemOne(t.Context(), Request{
		Model: "jev-latest",
		State: map[string]string{"review": "I was charged twice for my subscription and I am furious."},
		Questions: map[string]Question{
			"billing": Noul{Instructions: "Does the review mention billing?"},
			"topic": Choice{
				Instructions: "What is the review about?",
				Criteria:     map[string]string{"payment": "Charges, invoices, refunds", "shipping": "Delivery of goods"},
			},
			"mood": Score{
				Instructions: "How satisfied is the reviewer?",
				Criteria:     []string{"very unhappy", "neutral", "very happy"},
			},
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	if resp.Model == "" {
		t.Fatalf("model missing: %#v", resp)
	}
	if billing, ok := resp.Answers["billing"].(NoulAnswer); !ok || billing.Noul < 0.8 {
		t.Fatalf("billing: %#v", resp.Answers["billing"])
	}
	if topic, ok := resp.Answers["topic"].(ChoiceAnswer); !ok || topic.Choice != "payment" {
		t.Fatalf("topic: %#v", resp.Answers["topic"])
	}
	if _, ok := resp.Answers["mood"].(ScoreAnswer); !ok {
		t.Fatalf("mood: %#v", resp.Answers["mood"])
	}
}

func TestClientSystemOne_live_badKeyIsAPIError(t *testing.T) {
	t.Parallel()
	client := liveClient(t)
	client.APIKey = "not-a-real-key"
	_, err := client.SystemOne(t.Context(), Request{
		Model:     "jev-latest",
		State:     "",
		Questions: map[string]Question{"a": Noul{Instructions: "q"}},
	})
	if err == nil || !strings.HasPrefix(err.Error(), "api 401") {
		t.Fatalf("err: %v", err)
	}
}
