package shared

import "testing"

func TestSourceHelpers(t *testing.T) {
	t.Parallel()
	tokens := []Token{
		{Kind: Word, Text: "Draw", Span: Span{Start: Position{Offset: 0}, End: Position{Offset: 4}}},
		{Kind: Period, Text: ".", Span: Span{Start: Position{Offset: 4}, End: Position{Offset: 5}}},
	}
	if got := SpanOf(tokens); got != (Span{Start: Position{Offset: 0}, End: Position{Offset: 5}}) {
		t.Fatalf("span = %#v", got)
	}
	if got := SliceSpan("Draw.", SpanOf(tokens)); got != "Draw." {
		t.Fatalf("slice = %q", got)
	}
	if got := TopLevelIndex(tokens, Period); got != 1 {
		t.Fatalf("period index = %d", got)
	}
}
