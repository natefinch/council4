package parser

import (
	"testing"

	"github.com/natefinch/council4/cardgen/oracle/shared"
)

func TestNormalizedWords(t *testing.T) {
	t.Parallel()
	tokens := []shared.Token{
		{Kind: shared.Word, Text: "Draw"},
		{Kind: shared.Period, Text: "."},
		{Kind: shared.Word, Text: "GAIN"},
	}
	got := normalizedWords(tokens)
	if len(got) != 2 || got[0] != "draw" || got[1] != "gain" {
		t.Fatalf("normalizedWords = %#v, want [draw gain]", got)
	}
}
