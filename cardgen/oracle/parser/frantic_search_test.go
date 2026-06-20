package parser

import "testing"

func TestParseFranticSearchOrderedEffects(t *testing.T) {
	t.Parallel()
	document, diagnostics := Parse(
		"Draw two cards, then discard two cards. Untap up to three lands.",
		Context{InstantOrSorcery: true},
	)
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	sentences := document.Abilities[0].Sentences
	if len(sentences) != 2 || len(sentences[0].Effects) != 2 || len(sentences[1].Effects) != 1 {
		t.Fatalf("sentences = %#v, want draw/discard then untap", sentences)
	}
	untap := sentences[1].Effects[0]
	if untap.Kind != EffectUntap ||
		!untap.Exact ||
		!untap.Amount.RangeKnown ||
		untap.Amount.Minimum != 0 ||
		untap.Amount.Maximum != 3 ||
		untap.Selection.Kind != SelectionLand ||
		untap.Selection.Controller != SelectionControllerAny {
		t.Fatalf("untap = %#v, want exact controller choice of up to three lands", untap)
	}
}

func TestParseBoundedUntapBroadFormsExact(t *testing.T) {
	t.Parallel()
	cases := []struct {
		source     string
		maximum    int
		kind       SelectionKind
		controller SelectionController
	}{
		{"Untap up to two lands.", 2, SelectionLand, SelectionControllerAny},
		{"Untap up to four lands.", 4, SelectionLand, SelectionControllerAny},
		{"Untap up to three creatures.", 3, SelectionCreature, SelectionControllerAny},
		{"Untap up to two artifacts.", 2, SelectionArtifact, SelectionControllerAny},
		{"Untap up to two permanents.", 2, SelectionPermanent, SelectionControllerAny},
		{"Untap up to three lands you control.", 3, SelectionLand, SelectionControllerYou},
		{"Untap up to three lands an opponent controls.", 3, SelectionLand, SelectionControllerOpponent},
	}
	for _, tc := range cases {
		t.Run(tc.source, func(t *testing.T) {
			t.Parallel()
			document, diagnostics := Parse(tc.source, Context{InstantOrSorcery: true})
			if len(diagnostics) != 0 {
				t.Fatalf("diagnostics = %#v", diagnostics)
			}
			effects := document.Abilities[0].Sentences[0].Effects
			if len(effects) != 1 {
				t.Fatalf("effects = %#v, want one untap", effects)
			}
			untap := effects[0]
			if untap.Kind != EffectUntap ||
				!untap.Exact ||
				!untap.Amount.RangeKnown ||
				untap.Amount.Minimum != 0 ||
				untap.Amount.Maximum != tc.maximum ||
				untap.Selection.Kind != tc.kind ||
				untap.Selection.Controller != tc.controller {
				t.Fatalf("untap = %#v, want exact bounded untap", untap)
			}
		})
	}
}

func TestParseFranticSearchUntapNearMissesFailClosed(t *testing.T) {
	t.Parallel()
	for _, source := range []string{
		"Untap up to three tapped lands.",
		"Untap up to three random lands.",
		"Untap three lands.",
	} {
		t.Run(source, func(t *testing.T) {
			t.Parallel()
			document, diagnostics := Parse(source, Context{InstantOrSorcery: true})
			if len(diagnostics) != 0 {
				t.Fatalf("diagnostics = %#v", diagnostics)
			}
			effects := document.Abilities[0].Sentences[0].Effects
			if len(effects) != 1 || effects[0].Exact {
				t.Fatalf("effects = %#v, want one fail-closed untap", effects)
			}
		})
	}
}
