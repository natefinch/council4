package parser

import (
	"testing"

	"github.com/natefinch/council4/mtg/game/zone"
)

func TestParseCommanderFromCommandZonePut(t *testing.T) {
	t.Parallel()
	source := "{T}, Sacrifice this land: Put your commander into your hand from the command zone."
	document, diagnostics := Parse(source, Context{CardName: "Test"})
	if len(diagnostics) != 0 {
		t.Fatalf("Parse(%q) diagnostics = %#v", source, diagnostics)
	}
	if len(document.Abilities) != 1 || len(document.Abilities[0].Sentences) != 1 {
		t.Fatalf("Parse(%q) abilities = %#v", source, document.Abilities)
	}
	effects := document.Abilities[0].Sentences[0].Effects
	if len(effects) != 1 {
		t.Fatalf("Parse(%q) effects = %#v, want one", source, effects)
	}
	put := effects[0]
	if put.Kind != EffectPut {
		t.Fatalf("effect kind = %v, want EffectPut", put.Kind)
	}
	if put.Selection.Kind != SelectionCommander {
		t.Fatalf("selection kind = %v, want SelectionCommander", put.Selection.Kind)
	}
	if put.FromZone != zone.Command {
		t.Fatalf("FromZone = %v, want Command", put.FromZone)
	}
	if put.ToZone != zone.Hand {
		t.Fatalf("ToZone = %v, want Hand", put.ToZone)
	}
}

func TestCommandZonePhraseRecognition(t *testing.T) {
	t.Parallel()
	for _, tc := range []struct {
		source string
		want   zone.Type
	}{
		{"Put it into your hand from the command zone.", zone.Command},
		{"Put it into your hand from your command zone.", zone.Command},
		{"Put it into your hand from your graveyard.", zone.Graveyard},
	} {
		t.Run(tc.source, func(t *testing.T) {
			t.Parallel()
			document, diagnostics := Parse(tc.source, Context{InstantOrSorcery: true, CardName: "Test"})
			if len(diagnostics) != 0 {
				t.Fatalf("Parse(%q) diagnostics = %#v", tc.source, diagnostics)
			}
			effects := document.Abilities[0].Sentences[0].Effects
			if effects[0].FromZone != tc.want {
				t.Fatalf("FromZone = %v, want %v", effects[0].FromZone, tc.want)
			}
		})
	}
}
