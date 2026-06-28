package parser

import (
	"slices"
	"testing"

	"github.com/natefinch/council4/mtg/game/types"
)

func becomeTypeEffect(t *testing.T, name, text string) (EffectSyntax, bool) {
	t.Helper()
	doc, _ := Parse(text, Context{CardName: name})
	for a := range doc.Abilities {
		ability := &doc.Abilities[a]
		for s := range ability.Sentences {
			for e := range ability.Sentences[s].Effects {
				effect := ability.Sentences[s].Effects[e]
				if effect.Kind == EffectBecomeType {
					return effect, true
				}
			}
		}
	}
	return EffectSyntax{}, false
}

func TestParseBecomeTypeAddsTypeOnly(t *testing.T) {
	effect, ok := becomeTypeEffect(t, "Liquimetal Coating",
		"{T}: Target permanent becomes an artifact in addition to its other types until end of turn.")
	if !ok {
		t.Fatal("no become-type effect parsed")
	}
	if !effect.BecomeTypeUntilEndOfTurn {
		t.Error("expected until-end-of-turn duration")
	}
	if !slices.Equal(effect.BecomeTypeAddTypes, []types.Card{types.Artifact}) {
		t.Errorf("add types = %v, want [Artifact]", effect.BecomeTypeAddTypes)
	}
	if len(effect.BecomeTypeAddColors) != 0 {
		t.Errorf("add colors = %v, want none", effect.BecomeTypeAddColors)
	}
}

func TestParseBecomeTypeAddsColorAndType(t *testing.T) {
	effect, ok := becomeTypeEffect(t, "Unctus, Grand Metatect",
		"{U/P}: Until end of turn, target creature you control becomes a blue artifact in addition to its other colors and types.")
	if !ok {
		t.Fatal("no become-type effect parsed")
	}
	if !effect.BecomeTypeUntilEndOfTurn {
		t.Error("expected until-end-of-turn duration")
	}
	if !slices.Equal(effect.BecomeTypeAddTypes, []types.Card{types.Artifact}) {
		t.Errorf("add types = %v, want [Artifact]", effect.BecomeTypeAddTypes)
	}
	if !slices.Equal(effect.BecomeTypeAddColors, []Color{ColorBlue}) {
		t.Errorf("add colors = %v, want [ColorBlue]", effect.BecomeTypeAddColors)
	}
}

// TestParseReferencedTypeGrantPermanent asserts that the permanent
// referenced-object reanimation rider parses into an EffectBecomeType with no
// until-end-of-turn duration, the referenced-object context, and the added
// colors, card types, and creature subtypes.
func TestParseReferencedTypeGrantPermanent(t *testing.T) {
	effect, ok := becomeTypeEffect(t, "Rise from the Grave",
		"Put target creature card from a graveyard onto the battlefield under your control. That creature is a black Zombie in addition to its other colors and types.")
	if !ok {
		t.Fatal("no become-type effect parsed")
	}
	if effect.BecomeTypeUntilEndOfTurn {
		t.Error("expected no until-end-of-turn duration")
	}
	if effect.Context != EffectContextReferencedObject {
		t.Errorf("context = %v, want %v", effect.Context, EffectContextReferencedObject)
	}
	if !slices.Equal(effect.BecomeTypeAddColors, []Color{ColorBlack}) {
		t.Errorf("add colors = %v, want [ColorBlack]", effect.BecomeTypeAddColors)
	}
	if !slices.Equal(effect.BecomeTypeAddSubtypes, []types.Sub{types.Zombie}) {
		t.Errorf("add subtypes = %v, want [Zombie]", effect.BecomeTypeAddSubtypes)
	}
	if len(effect.BecomeTypeAddTypes) != 0 {
		t.Errorf("add types = %v, want none", effect.BecomeTypeAddTypes)
	}
}

// TestParseReferencedTypeGrantSubjects asserts the back-reference subjects and
// suffixes the permanent reanimation rider accepts ("It's", "It is", "The
// creature is", and the "creature types" suffix) each parse a subtype grant.
func TestParseReferencedTypeGrantSubjects(t *testing.T) {
	for _, text := range []string{
		"Put target creature card from a graveyard onto the battlefield under your control. It's a Phyrexian in addition to its other types.",
		"Return target creature card from your graveyard to the battlefield. It is a Vampire in addition to its other types.",
		"Return target creature card from your graveyard to the battlefield. The creature is a Skeleton in addition to its other creature types.",
	} {
		effect, ok := becomeTypeEffect(t, "Test", text)
		if !ok {
			t.Fatalf("no become-type effect parsed for %q", text)
		}
		if effect.BecomeTypeUntilEndOfTurn {
			t.Errorf("expected no until-end-of-turn duration for %q", text)
		}
		if len(effect.BecomeTypeAddSubtypes) != 1 {
			t.Errorf("add subtypes = %v, want one for %q", effect.BecomeTypeAddSubtypes, text)
		}
	}
}

// TestParseReferencedTypeGrantFailsClosed asserts that an until-end-of-turn
// duration, an unrecognized type word, and a compound "is black and is" subject
// all fail closed rather than parsing a permanent referenced-object grant.
func TestParseReferencedTypeGrantFailsClosed(t *testing.T) {
	for _, text := range []string{
		"Return target creature card from your graveyard to the battlefield. That creature is a Zombie in addition to its other types until end of turn.",
		"Return target creature card from your graveyard to the battlefield. That creature is a Wizmacallit in addition to its other types.",
		"Put target creature card from a graveyard onto the battlefield under your control. That creature is black and is a Nightmare in addition to its other creature types.",
	} {
		if _, ok := becomeTypeEffect(t, "Test", text); ok {
			t.Errorf("expected no become-type effect for %q", text)
		}
	}
}

// TestParseBecomeTypeColorAndTypeFailsClosed asserts that mismatched additive
// wordings fail closed: a color word with only "in addition to its other types"
// (missing "colors and"), and the "colors and types" tail with no color word.
func TestParseBecomeTypeColorAndTypeFailsClosed(t *testing.T) {
	for _, text := range []string{
		"{U/P}: Until end of turn, target creature you control becomes a blue artifact in addition to its other types.",
		"{U/P}: Until end of turn, target creature you control becomes an artifact in addition to its other colors and types.",
		"{2}: Until end of turn, target nonartifact creature gets +1/+0 and becomes an artifact in addition to its other types.",
	} {
		if _, ok := becomeTypeEffect(t, "Test", text); ok {
			t.Errorf("expected no become-type effect for %q", text)
		}
	}
}
