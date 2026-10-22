package typesafe

import (
	"strings"
	"testing"

	"hegel.dev/go/hegel"
)

func TestTruncateBody_neverExceedsLimit(t *testing.T) {
	hegel.Test(t, func(ht *hegel.T) {
		n := hegel.Draw(ht, hegel.Integers(0, maxErrorBody*3))
		body := strings.Repeat("x", n)
		got := truncateBody([]byte(body))
		if len(got) > maxErrorBody+len("…") {
			ht.Fatalf("len %d body input %d", len(got), n)
		}
		if n <= maxErrorBody && got != body {
			ht.Fatalf("short body should be unchanged: %q", got)
		}
		if n > maxErrorBody && !strings.HasSuffix(got, "…") {
			ht.Fatalf("expected ellipsis suffix: %q", got)
		}
	})
}
