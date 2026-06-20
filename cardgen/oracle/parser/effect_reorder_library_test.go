package parser

import (
	"testing"

	"github.com/natefinch/council4/cardgen/oracle/shared"
)

func TestParseExactLibraryTopReorderEffect(t *testing.T) {
	t.Parallel()
	source := "Look at the top three cards of your library, then put them back in any order."
	document, diagnostics := Parse(source, Context{InstantOrSorcery: true})
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	effects := document.Abilities[0].Sentences[0].Effects
	if len(effects) != 1 {
		t.Fatalf("effects = %#v, want one library reorder", effects)
	}
	effect := effects[0]
	if effect.Kind != EffectReorderLibraryTop ||
		effect.Context != EffectContextController ||
		!effect.Exact ||
		!effect.Amount.Known ||
		effect.Amount.Value != 3 {
		t.Fatalf("effect = %#v, want exact controller reorder of three", effect)
	}
	for name, spanWant := range map[string]struct {
		span shared.Span
		want string
	}{
		"effect": {span: effect.Span, want: source},
		"clause": {span: effect.ClauseSpan, want: source},
		"verb":   {span: effect.VerbSpan, want: "Look"},
		"amount": {span: effect.Amount.Span, want: "three"},
	} {
		if got := shared.SliceSpan(source, spanWant.span); got != spanWant.want {
			t.Errorf("%s span = %q, want %q", name, got, spanWant.want)
		}
	}
}

func TestLibraryTopReorderBoundaryFailsClosed(t *testing.T) {
	t.Parallel()
	for _, source := range []string{
		"Look at the bottom three cards of your library, then put them back in any order.",
		"Look at the top X cards of your library, then put them back in any order.",
		"Look at the top three cards of an opponent's library, then put them back in any order.",
		"Look at the top three cards of your library, then put one back in any order.",
		"Look at the top three cards of your library, then put them back in a random order.",
		"Look at the top three cards of your library, then put them back in the same order.",
		"Look at the top three cards of your library. Put them back in any order.",
	} {
		t.Run(source, func(t *testing.T) {
			t.Parallel()
			document, _ := Parse(source, Context{InstantOrSorcery: true})
			for _, ability := range document.Abilities {
				for _, sentence := range ability.Sentences {
					for _, effect := range sentence.Effects {
						if effect.Kind == EffectReorderLibraryTop {
							t.Fatalf("%q unexpectedly recognized as exact library reorder: %#v", source, effect)
						}
					}
				}
			}
		})
	}
}

func TestParseExactOptionalControllerShuffle(t *testing.T) {
	t.Parallel()
	for _, source := range []string{"You may shuffle.", "You may shuffle your library."} {
		t.Run(source, func(t *testing.T) {
			t.Parallel()
			document, diagnostics := Parse(source, Context{InstantOrSorcery: true})
			if len(diagnostics) != 0 {
				t.Fatalf("diagnostics = %#v", diagnostics)
			}
			effect := document.Abilities[0].Sentences[0].Effects[0]
			if effect.Kind != EffectShuffle || !effect.Optional || !effect.Exact {
				t.Fatalf("effect = %#v, want exact optional controller shuffle", effect)
			}
			if got := shared.SliceSpan(source, effect.OptionalSpan); got != "You may" {
				t.Fatalf("optional span = %q, want %q", got, "You may")
			}
		})
	}
}
