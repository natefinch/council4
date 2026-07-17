package parser

import (
	"slices"
	"testing"

	"github.com/natefinch/council4/mtg/game/types"
)

// TestParseGroupControlledBecomeTypeAddsSubtype asserts that the resolving group
// rider a reanimation spell applies to its controller's creatures ("Then each
// creature you control becomes a Phyrexian in addition to its other types.",
// Breach the Multiverse) parses into an EffectBecomeType carrying the
// controlled-creature static subject, the added subtype, and no until-end-of-turn
// duration.
func TestParseGroupControlledBecomeTypeAddsSubtype(t *testing.T) {
	effect, ok := becomeTypeEffect(t, "Breach the Multiverse",
		"Then each creature you control becomes a Phyrexian in addition to its other types.")
	if !ok {
		t.Fatal("no become-type effect parsed")
	}
	if effect.StaticSubject.Kind != EffectStaticSubjectControlledCreatures {
		t.Errorf("static subject = %v, want %v", effect.StaticSubject.Kind, EffectStaticSubjectControlledCreatures)
	}
	if effect.BecomeTypeUntilEndOfTurn {
		t.Error("expected no until-end-of-turn duration")
	}
	if !slices.Equal(effect.BecomeTypeAddSubtypes, []types.Sub{types.Phyrexian}) {
		t.Errorf("add subtypes = %v, want [Phyrexian]", effect.BecomeTypeAddSubtypes)
	}
	if len(effect.BecomeTypeAddTypes) != 0 || len(effect.BecomeTypeAddColors) != 0 {
		t.Errorf("add types = %v, add colors = %v, want none", effect.BecomeTypeAddTypes, effect.BecomeTypeAddColors)
	}
}

// TestParseGroupControlledBecomeTypeAddsColorAndSubtype asserts the group rider
// accepts the color-and-type additive tail without the leading "Then" connective
// and records both the added color and the added subtype.
func TestParseGroupControlledBecomeTypeAddsColorAndSubtype(t *testing.T) {
	effect, ok := becomeTypeEffect(t, "Test",
		"Each creature you control becomes a black Zombie in addition to its other colors and types.")
	if !ok {
		t.Fatal("no become-type effect parsed")
	}
	if effect.StaticSubject.Kind != EffectStaticSubjectControlledCreatures {
		t.Errorf("static subject = %v, want %v", effect.StaticSubject.Kind, EffectStaticSubjectControlledCreatures)
	}
	if !slices.Equal(effect.BecomeTypeAddColors, []Color{ColorBlack}) {
		t.Errorf("add colors = %v, want [ColorBlack]", effect.BecomeTypeAddColors)
	}
	if !slices.Equal(effect.BecomeTypeAddSubtypes, []types.Sub{types.Zombie}) {
		t.Errorf("add subtypes = %v, want [Zombie]", effect.BecomeTypeAddSubtypes)
	}
}

// TestParseGroupControlledBecomeTypeFailsClosed asserts the group rider fails
// closed on near misses: a missing additive "in addition to its other types"
// tail (a type-setting change, not an additive grant), an until-end-of-turn
// duration, a subject that is not "each creature you control" (a different
// controller or a non-creature permanent), and an unrecognized type word. None
// may parse into a group become-type effect.
func TestParseGroupControlledBecomeTypeFailsClosed(t *testing.T) {
	for _, text := range []string{
		"Each creature you control becomes a Phyrexian.",
		"Each creature you control becomes a Phyrexian in addition to its other types until end of turn.",
		"Each creature an opponent controls becomes a Phyrexian in addition to its other types.",
		"Each artifact you control becomes a Phyrexian in addition to its other types.",
		"Each creature you control becomes a Wizmacallit in addition to its other types.",
	} {
		effect, ok := becomeTypeEffect(t, "Test", text)
		if ok && effect.StaticSubject.Kind == EffectStaticSubjectControlledCreatures {
			t.Errorf("expected no group become-type effect for %q", text)
		}
	}
}
