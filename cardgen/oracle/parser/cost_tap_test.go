package parser

import (
	"testing"

	"github.com/natefinch/council4/mtg/game/types"
)

func TestParseTapPermanentsSubtypeFamilies(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name    string
		source  string
		amount  int
		subtype types.Sub
	}{
		{"land gate subtype", "Tap an untapped Gate you control: Draw a card.", 1, types.Gate},
		{"land desert subtype", "Tap an untapped Desert you control: You gain 1 life.", 1, types.Desert},
		{"creature subtype", "Tap two untapped Elves you control: Draw a card.", 2, types.Elf},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			component := soleCostComponent(t, test.source)
			if component.Kind != CostComponentTapPermanents {
				t.Fatalf("kind = %v, want tap permanents", component.Kind)
			}
			if !component.AmountKnown || component.AmountValue != test.amount {
				t.Fatalf("amount = (%d, %v), want %d", component.AmountValue, component.AmountKnown, test.amount)
			}
			if !component.RequireUntapped || component.ObjectController != ControllerRelationYouControl {
				t.Fatalf("component = %#v, want untapped you-control", component)
			}
			if len(component.SubtypesAny) != 1 || component.SubtypesAny[0] != test.subtype {
				t.Fatalf("subtypes = %v, want [%s]", component.SubtypesAny, test.subtype)
			}
		})
	}
}

func TestParseTapPermanentsSupertype(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name      string
		source    string
		amount    int
		noun      ObjectNoun
		supertype types.Super
	}{
		{"legendary creature", "Tap an untapped legendary creature you control: Add one mana of any color.", 1, ObjectNounCreature, types.Legendary},
		{"legendary artifact", "Tap two untapped legendary artifacts you control: Draw a card.", 2, ObjectNounArtifact, types.Legendary},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			component := soleCostComponent(t, test.source)
			if component.Kind != CostComponentTapPermanents {
				t.Fatalf("kind = %v, want tap permanents", component.Kind)
			}
			if !component.AmountKnown || component.AmountValue != test.amount {
				t.Fatalf("amount = (%d, %v), want %d", component.AmountValue, component.AmountKnown, test.amount)
			}
			if !component.RequireUntapped || component.ObjectController != ControllerRelationYouControl {
				t.Fatalf("component = %#v, want untapped you-control", component)
			}
			if component.ObjectNoun != test.noun {
				t.Fatalf("noun = %v, want %v", component.ObjectNoun, test.noun)
			}
			if !component.SupertypeKnown || component.ObjectSupertype != test.supertype {
				t.Fatalf("supertype = (%v, %v), want %v", component.ObjectSupertype, component.SupertypeKnown, test.supertype)
			}
		})
	}
}

func TestParseTapPermanentsTwoTypeUnion(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name   string
		source string
		amount int
		first  ObjectNoun
		second ObjectNoun
	}{
		{"and/or union", "Tap two untapped artifacts and/or creatures you control: Draw a card.", 2, ObjectNounArtifact, ObjectNounCreature},
		{"or union", "Tap an untapped artifact or creature you control: Add {C}.", 1, ObjectNounArtifact, ObjectNounCreature},
		{"creature land union", "Tap two untapped creatures and/or lands you control: Draw a card.", 2, ObjectNounCreature, ObjectNounLand},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			component := soleCostComponent(t, test.source)
			if component.Kind != CostComponentTapPermanents {
				t.Fatalf("kind = %v, want tap permanents", component.Kind)
			}
			if !component.AmountKnown || component.AmountValue != test.amount {
				t.Fatalf("amount = (%d, %v), want %d", component.AmountValue, component.AmountKnown, test.amount)
			}
			if !component.RequireUntapped || component.ObjectController != ControllerRelationYouControl {
				t.Fatalf("component = %#v, want untapped you-control", component)
			}
			if component.ObjectNoun != test.first || component.SecondObjectNoun != test.second {
				t.Fatalf("nouns = (%v, %v), want (%v, %v)",
					component.ObjectNoun, component.SecondObjectNoun, test.first, test.second)
			}
		})
	}
}

func TestParseTapPermanentsUnionFailsClosed(t *testing.T) {
	t.Parallel()
	for _, source := range []string{
		// "permanent" carries no constraining union member.
		"Tap two untapped artifacts and/or permanents you control: Draw a card.",
		// A repeated type is not a union.
		"Tap two untapped creatures and/or creatures you control: Draw a card.",
		// A subtype is not a permanent-type union member here.
		"Tap two untapped artifacts and/or Goblins you control: Draw a card.",
	} {
		component := soleCostComponent(t, source)
		if component.RequireUntapped || component.SecondObjectNoun != ObjectNounUnknown {
			t.Fatalf("%q: component = %#v, want bare unrecognized tap", source, component)
		}
	}
}
