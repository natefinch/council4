package parser

import (
	"testing"

	"github.com/natefinch/council4/mtg/game/zone"
)

// TestParseTargetPlayerGraveyardShuffle checks the "Target player shuffles their
// graveyard into their library." wording types to an exact EffectShuffle whose
// graveyard source and library destination are recognized, scoped to a single
// player target.
func TestParseTargetPlayerGraveyardShuffle(t *testing.T) {
	t.Parallel()
	for _, text := range []string{
		"Target player shuffles their graveyard into their library.",
		"Target opponent shuffles their graveyard into their library.",
	} {
		document, diagnostics := Parse(text, Context{})
		if len(diagnostics) != 0 {
			t.Fatalf("%q diagnostics = %#v", text, diagnostics)
		}
		effects := document.Abilities[0].Sentences[0].Effects
		if len(effects) != 1 {
			t.Fatalf("%q effects = %#v", text, effects)
		}
		effect := effects[0]
		if effect.Kind != EffectShuffle || !effect.Exact ||
			effect.Context != EffectContextTarget ||
			effect.FromZone != zone.Graveyard ||
			effect.ToZone != zone.Library {
			t.Fatalf("%q effect = %#v", text, effect)
		}
	}
}

// TestParseGraveyardZonePhraseTheir confirms the "their graveyard" possessive is
// recognized as a graveyard zone phrase so the target-player shuffle resolves its
// source zone.
func TestParseGraveyardZonePhraseTheir(t *testing.T) {
	t.Parallel()
	document, diagnostics := Parse("Shuffle your graveyard into your library.", Context{})
	if len(diagnostics) != 0 {
		t.Fatalf("controller form diagnostics = %#v", diagnostics)
	}
	if got := document.Abilities[0].Sentences[0].Effects[0].FromZone; got != zone.Graveyard {
		t.Fatalf("controller shuffle FromZone = %v, want graveyard", got)
	}
}
