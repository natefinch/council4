package parser

import (
	"testing"

	"github.com/natefinch/council4/mtg/game/zone"
)

// firstPutEffect parses a single "Put ..." sentence and returns its lone effect,
// requiring the parse to be diagnostic-free and shaped as one EffectPut.
func firstPutEffect(t *testing.T, source string) EffectSyntax {
	t.Helper()
	document, diagnostics := Parse(source, Context{InstantOrSorcery: true, CardName: "Test"})
	if len(diagnostics) != 0 {
		t.Fatalf("Parse(%q) diagnostics = %#v", source, diagnostics)
	}
	if len(document.Abilities) != 1 || len(document.Abilities[0].Sentences) != 1 {
		t.Fatalf("Parse(%q) shape = %#v", source, document.Abilities)
	}
	effects := document.Abilities[0].Sentences[0].Effects
	if len(effects) != 1 || effects[0].Kind != EffectPut {
		t.Fatalf("Parse(%q) effects = %#v", source, effects)
	}
	return effects[0]
}

// TestParsePutTargetPermanentOnLibraryIsExact verifies the in-play permanent
// tuck "Put target <permanent> on top/bottom of its owner's library." round-trips
// to an exact EffectPut carrying the library destination and the exact single
// permanent target, so the lowerer can trust the reconstruction.
func TestParsePutTargetPermanentOnLibraryIsExact(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name       string
		source     string
		wantBottom bool
	}{
		{"creature top", "Put target creature on top of its owner's library.", false},
		{"creature the top", "Put target creature on the top of its owner's library.", false},
		{"creature bottom", "Put target creature on the bottom of its owner's library.", true},
		{"land top", "Put target land on top of its owner's library.", false},
		{"nonland permanent top", "Put target nonland permanent on top of its owner's library.", false},
		{"type union top", "Put target artifact or enchantment on top of its owner's library.", false},
		{"power qualifier bottom", "Put target creature with power 4 or greater on the bottom of its owner's library.", true},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			effect := firstPutEffect(t, test.source)
			if !effect.Exact {
				t.Fatalf("effect Exact = false, want true: %#v", effect)
			}
			if effect.ToZone != zone.Library {
				t.Fatalf("effect ToZone = %v, want Library", effect.ToZone)
			}
			wantDest := EffectDestinationTop
			if test.wantBottom {
				wantDest = EffectDestinationBottom
			}
			if effect.Destination != wantDest {
				t.Fatalf("effect Destination = %v, want %v", effect.Destination, wantDest)
			}
			if len(effect.Targets) != 1 || !effect.Targets[0].Exact {
				t.Fatalf("effect targets = %#v, want one exact target", effect.Targets)
			}
		})
	}
}

// TestParsePutNonTargetPermanentOnLibraryNotExact verifies a non-target subject
// tuck ("Put a creature you control on top of its owner's library.", Nulltread
// Gargantuan) does not round-trip through the single-target recognizer, so it
// stays inexact and the lowerer fails closed rather than moving the wrong object.
func TestParsePutNonTargetPermanentOnLibraryNotExact(t *testing.T) {
	t.Parallel()
	effect := firstPutEffect(t, "Put a creature you control on top of its owner's library.")
	if len(effect.Targets) != 0 {
		t.Fatalf("effect targets = %#v, want none for a non-target subject", effect.Targets)
	}
	if effect.Exact {
		t.Fatal("effect Exact = true, want false for the non-target tuck")
	}
}
