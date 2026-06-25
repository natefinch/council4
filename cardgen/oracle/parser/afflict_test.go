package parser

import "testing"

func TestExpandAfflictKeyword(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name   string
		source string
		want   string
	}{
		{
			"wildfire eternal",
			"Afflict 4 (Whenever this creature becomes blocked, defending player loses 4 life.)",
			"Whenever this creature becomes blocked, defending player loses 4 life.",
		},
		{
			"bare keyword one",
			"Afflict 1",
			"Whenever this creature becomes blocked, defending player loses 1 life.",
		},
		{
			"bare keyword three",
			"Afflict 3",
			"Whenever this creature becomes blocked, defending player loses 3 life.",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			if got := expandAfflictKeyword(test.source); got != test.want {
				t.Fatalf("expandAfflictKeyword = %q, want %q", got, test.want)
			}
		})
	}
}

func TestExpandAfflictKeywordLeavesOtherTextAlone(t *testing.T) {
	t.Parallel()
	// The granted form ("creatures you control have afflict 2") is not a
	// standalone keyword line and must not be rewritten.
	granted := "Sliver creatures you control have afflict 2."
	if got := expandAfflictKeyword(granted); got != granted {
		t.Fatalf("rewrote granted afflict: %q", got)
	}
	if got := expandAfflictKeyword("Afflict"); got != "Afflict" {
		t.Fatalf("rewrote rankless keyword: %q", got)
	}
	if got := expandAfflictKeyword("Whenever Afflict attacks, draw a card."); got != "Whenever Afflict attacks, draw a card." {
		t.Fatalf("rewrote unrelated line: %q", got)
	}
}
