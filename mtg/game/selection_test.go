package game

import (
	"strings"
	"testing"

	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/types"
)

func TestSelectionRejectsTokenContradiction(t *testing.T) {
	problems := (Selection{NonToken: true, TokenOnly: true}).Validate()
	if len(problems) != 1 {
		t.Fatalf("problems = %#v, want one token contradiction", problems)
	}
}

func TestSelectionValidatesAnyOfAlternatives(t *testing.T) {
	t.Parallel()

	problems := (Selection{AnyOf: []Selection{{
		RequiredTypes: []types.Card{types.Land},
		ExcludedTypes: []types.Card{types.Land},
	}}}).Validate()
	if len(problems) != 1 || !strings.Contains(problems[0], "alternative 0") {
		t.Fatalf("problems = %#v, want nested alternative path", problems)
	}
}

func TestSelectionRejectsColorCardinalityContradictions(t *testing.T) {
	tests := []struct {
		name      string
		selection Selection
	}{
		{"colorless multicolored", Selection{Colorless: true, Multicolored: true}},
		{"colorless colored", Selection{Colorless: true, ColorsAny: []color.Color{color.Red}}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			problems := tt.selection.Validate()
			if len(problems) != 1 {
				t.Fatalf("problems = %#v, want one color-cardinality contradiction", problems)
			}
		})
	}
}

// TestSelectionEmptyDetectsManaValueLessThanEventPermanent ensures the
// event-relative "with lesser mana value" bound counts as a non-empty selection
// so a target carrying only that predicate still routes through the runtime
// matcher rather than being skipped as an empty filter.
func TestSelectionEmptyDetectsManaValueLessThanEventPermanent(t *testing.T) {
	if !(Selection{}).Empty() {
		t.Fatal("a zero Selection must be empty")
	}
	if (Selection{ManaValueLessThanEventPermanent: true}).Empty() {
		t.Fatal("a Selection carrying only ManaValueLessThanEventPermanent must not be empty")
	}
}
