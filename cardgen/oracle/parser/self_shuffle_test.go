package parser

import (
	"testing"

	"github.com/natefinch/council4/mtg/game/zone"
)

// TestParseSelfShuffleIntoOwnerLibrary checks that the dies / put-into-graveyard
// self-recursion "shuffle it into its owner's library" (and the optional "you
// may" form) types to an exact EffectShuffle whose destination is the library.
func TestParseSelfShuffleIntoOwnerLibrary(t *testing.T) {
	t.Parallel()
	for _, text := range []string{
		"When this creature dies, shuffle it into its owner's library.",
		"When this creature dies, you may shuffle it into its owner's library.",
	} {
		document, diagnostics := Parse(text, Context{})
		if len(diagnostics) != 0 {
			t.Fatalf("%q diagnostics = %#v", text, diagnostics)
		}
		effect := document.Abilities[0].Sentences[0].Effects[0]
		if effect.Kind != EffectShuffle || !effect.Exact ||
			effect.Context != EffectContextController ||
			effect.ToZone != zone.Library {
			t.Fatalf("%q effect = %#v", text, effect)
		}
	}
}
