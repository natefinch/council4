package game

import (
	"reflect"
	"testing"

	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/types"
)

func TestTargetPredicateSelectionMapsSubtypesAndSupertypes(t *testing.T) {
	predicate := TargetPredicate{
		PermanentTypes: []types.Card{types.Creature},
		Supertypes:     []types.Super{types.Legendary},
		Subtypes:       []types.Sub{types.Sub("Beast")},
		Controller:     ControllerYou,
		Another:        true,
	}
	selection := predicate.Selection()
	if !reflect.DeepEqual(selection.Supertypes, predicate.Supertypes) {
		t.Fatalf("supertypes = %#v, want %#v", selection.Supertypes, predicate.Supertypes)
	}
	if !reflect.DeepEqual(selection.SubtypesAny, predicate.Subtypes) {
		t.Fatalf("subtypesAny = %#v, want %#v", selection.SubtypesAny, predicate.Subtypes)
	}
	if !selection.ExcludeSource {
		t.Fatal("excludeSource = false, want true (Another must exclude the source)")
	}
}

func TestSelectionRejectsTokenContradiction(t *testing.T) {
	problems := (Selection{NonToken: true, TokenOnly: true}).Validate()
	if len(problems) != 1 {
		t.Fatalf("problems = %#v, want one token contradiction", problems)
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
