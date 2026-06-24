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
