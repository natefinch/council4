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
