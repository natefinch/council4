package parser

import (
	"testing"

	"github.com/natefinch/council4/mtg/game/zone"
)

func TestParseEachPlayerGraveyardShuffle(t *testing.T) {
	t.Parallel()
	document, diagnostics := Parse("Each player shuffles their graveyard into their library.", Context{})
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	effect := document.Abilities[0].Sentences[0].Effects[0]
	if effect.Kind != EffectShuffle ||
		!effect.Exact ||
		effect.Context != EffectContextEachPlayer ||
		effect.FromZone != zone.Graveyard ||
		effect.ToZone != zone.Library ||
		!effect.ShuffleEachPlayerGraveyardIntoLibrary {
		t.Fatalf("effect = %#v", effect)
	}
}
