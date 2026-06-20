package parser

import (
	"testing"

	"github.com/natefinch/council4/mtg/game/zone"
)

// TestParsePutSourceOnOwnersLibrary verifies the parser recognizes the
// destination of "put this [permanent] on top/bottom of its/their owner's
// library" as an EffectPut targeting the library with the typed destination,
// so the text-blind lowerer can route it without re-reading the wording.
func TestParsePutSourceOnOwnersLibrary(t *testing.T) {
	t.Parallel()
	for _, test := range []struct {
		source string
		want   EffectDestinationPosition
	}{
		{"Put this artifact on top of its owner's library.", EffectDestinationTop},
		{"Put this creature on the bottom of its owner's library.", EffectDestinationBottom},
		{"Put this permanent on top of their owner's library.", EffectDestinationTop},
	} {
		t.Run(test.source, func(t *testing.T) {
			t.Parallel()
			document, diagnostics := Parse(test.source, Context{InstantOrSorcery: true})
			if len(diagnostics) != 0 {
				t.Fatalf("diagnostics = %#v", diagnostics)
			}
			effect := document.Abilities[0].Sentences[0].Effects[0]
			if effect.Kind != EffectPut ||
				effect.ToZone != zone.Library ||
				effect.Destination != test.want {
				t.Fatalf("effect = %#v, want EffectPut to library with destination %v", effect, test.want)
			}
		})
	}
}
