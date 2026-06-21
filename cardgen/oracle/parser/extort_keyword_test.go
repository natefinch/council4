package parser

import (
	"slices"
	"testing"
)

func TestExpandExtortKeyword(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name   string
		source string
	}{
		{"bare keyword", "Extort"},
		{"with reminder", "Extort (Whenever you cast a spell, you may pay {W/B}. If you do, each opponent loses 1 life and you gain that much life.)"},
		{"after another keyword", "Flying\nExtort"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			got := expandExtortKeyword(test.source)
			if !containsLine(got, extortCanonicalText) {
				t.Fatalf("expandExtortKeyword(%q) = %q, want a line equal to the canonical text", test.source, got)
			}
		})
	}
}

func TestExpandExtortKeywordLeavesOtherTextAlone(t *testing.T) {
	t.Parallel()
	// A line that only mentions the word elsewhere must not be rewritten.
	if got := expandExtortKeyword("Whenever you Extort a tax, draw a card."); got != "Whenever you Extort a tax, draw a card." {
		t.Fatalf("rewrote unrelated line: %q", got)
	}
	if got := expandExtortKeyword("Extort, then draw a card."); got != "Extort, then draw a card." {
		t.Fatalf("rewrote keyword paired with other rules text: %q", got)
	}
}

func containsLine(source, line string) bool {
	return slices.Contains(splitSourceLines(source), line)
}

func splitSourceLines(source string) []string {
	var lines []string
	start := 0
	for i := 0; i < len(source); i++ {
		if source[i] == '\n' {
			lines = append(lines, source[start:i])
			start = i + 1
		}
	}
	return append(lines, source[start:])
}
