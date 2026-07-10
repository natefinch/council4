package parser

import (
	"testing"
)

// TestParseLookAtTargetPlayerLibrary checks that "Look at the top card of target
// player's library." types to an exact target-scoped EffectLookAtLibraryTop with
// a single player target, and that the controller "your library" form stays
// target-free.
func TestParseLookAtTargetPlayerLibrary(t *testing.T) {
	t.Parallel()
	cases := []struct {
		text       string
		targetText string
	}{
		{"Look at the top card of target player's library.", "target player"},
		{"Look at the top card of target opponent's library.", "target opponent"},
	}
	for _, tc := range cases {
		document, diagnostics := Parse(tc.text, Context{})
		if len(diagnostics) != 0 {
			t.Fatalf("%q diagnostics = %#v", tc.text, diagnostics)
		}
		effect := document.Abilities[0].Sentences[0].Effects[0]
		if effect.Kind != EffectLookAtLibraryTop || !effect.Exact ||
			effect.Context != EffectContextTarget ||
			len(effect.Targets) != 1 {
			t.Fatalf("%q effect = %#v", tc.text, effect)
		}
		if got := effect.Targets[0].Text; got != tc.targetText {
			t.Fatalf("%q target text = %q, want retyped %q", tc.text, got, tc.targetText)
		}
	}
}

// TestParseLookAtYourLibraryStaysControllerScoped confirms the controller
// "your library" peek is still classified as EffectLookAtLibraryTop and carries
// no target, unaffected by the target-player retyping.
func TestParseLookAtYourLibraryStaysControllerScoped(t *testing.T) {
	t.Parallel()
	document, _ := Parse("Look at the top card of your library.", Context{})
	effect := document.Abilities[0].Sentences[0].Effects[0]
	if effect.Kind != EffectLookAtLibraryTop || len(effect.Targets) != 0 {
		t.Fatalf("effect = %#v, want controller-scoped look with no target", effect)
	}
}
