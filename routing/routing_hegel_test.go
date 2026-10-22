package routing

import (
	"strings"
	"testing"

	"hegel.dev/go/hegel"

	"github.com/havriil/jevpipe/typesafe"
)

func TestScalarCandidates_includesScalarLeaves(t *testing.T) {
	hegel.Test(t, func(ht *hegel.T) {
		key := hegel.Draw(ht, hegel.Text().Alphabet("abc").MinSize(1).MaxSize(8))
		n := hegel.Draw(ht, hegel.Integers(-500, 500))
		state := map[string]any{key: n}
		want := scalarString(n)
		found := false
		for _, c := range scalarCandidates(state) {
			if c.Value == want && c.Path == key {
				found = true
				break
			}
		}
		if !found {
			ht.Fatalf("missing candidate key=%q value=%q in %#v", key, want, scalarCandidates(state))
		}
	})
}

func TestQuestion_yesNoPromptsUseNoul(t *testing.T) {
	hegel.Test(t, func(ht *hegel.T) {
		prefix := hegel.Draw(ht, hegel.SampledFrom([]string{
			"is", "Is", "does", "Does", "has", "Has",
		}))
		tail := hegel.Draw(ht, hegel.Text().Alphabet("abcdefghijklmnopqrstuvwxyz ").MinSize(0).MaxSize(40))
		prompt := prefix + " " + strings.TrimSpace(tail)
		state := map[string]any{"x": hegel.Draw(ht, hegel.Integers(1, 99))}
		q := Question(state, prompt)
		if q.Type() != typesafe.TypeNoul {
			ht.Fatalf("prompt %q: got %v", prompt, q.Type())
		}
	})
}
