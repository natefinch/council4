package parser

import (
	"testing"

	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/zone"
)

func onlyEffect(t *testing.T, source string) EffectSyntax {
	t.Helper()
	document, diagnostics := Parse(source, Context{InstantOrSorcery: true})
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v, want none", diagnostics)
	}
	if len(document.Abilities) != 1 || len(document.Abilities[0].Sentences) != 1 {
		t.Fatalf("abilities = %#v, want exactly one single-sentence ability", document.Abilities)
	}
	effects := document.Abilities[0].Sentences[0].Effects
	if len(effects) != 1 {
		t.Fatalf("effects = %#v, want exactly one", effects)
	}
	return effects[0]
}

// TestParseReturnExiledCardsWithCounter proves the recognizer captures "Put all
// exiled cards you own with intel counters on them into your hand." as an exact
// controller-scoped return from exile to hand, carrying the named marker counter
// it reads text-blind.
func TestParseReturnExiledCardsWithCounter(t *testing.T) {
	effect := onlyEffect(t, "Put all exiled cards you own with intel counters on them into your hand.")

	if effect.Kind != EffectReturnExiledCardsWithCounter {
		t.Fatalf("effect kind = %v, want EffectReturnExiledCardsWithCounter", effect.Kind)
	}
	if !effect.Exact {
		t.Fatal("effect should be exact")
	}
	if effect.Context != EffectContextController {
		t.Fatalf("context = %v, want controller", effect.Context)
	}
	if effect.FromZone != zone.Exile {
		t.Fatalf("from zone = %v, want Exile", effect.FromZone)
	}
	if effect.ToZone != zone.Hand {
		t.Fatalf("to zone = %v, want Hand", effect.ToZone)
	}
	if !effect.CounterKnown || effect.CounterKind != counter.Intel {
		t.Fatalf("counter = (known=%v, kind=%v), want (true, intel)", effect.CounterKnown, effect.CounterKind)
	}
}

// TestParseReturnExiledCardsWithCounterIsCounterAgnostic proves the recognizer
// reads whichever named marker counter the wording uses rather than a hard-coded
// name, so every named-counter-exile card benefits.
func TestParseReturnExiledCardsWithCounterIsCounterAgnostic(t *testing.T) {
	for _, tc := range []struct {
		word string
		kind counter.Kind
	}{
		{"void", counter.Void},
		{"collection", counter.Collection},
		{"croak", counter.Croak},
	} {
		effect := onlyEffect(t, "Put all exiled cards you own with "+tc.word+" counters on them into your hand.")
		if effect.Kind != EffectReturnExiledCardsWithCounter {
			t.Fatalf("%q: effect kind = %v, want EffectReturnExiledCardsWithCounter", tc.word, effect.Kind)
		}
		if !effect.CounterKnown || effect.CounterKind != tc.kind {
			t.Fatalf("%q: counter = (known=%v, kind=%v), want (true, %v)", tc.word, effect.CounterKnown, effect.CounterKind, tc.kind)
		}
	}
}

// TestParseReturnExiledCardsWithCounterFailsClosed proves wordings the recognizer
// does not fully model do not become an exact return effect: they fall through to
// the generic parser (kind != EffectReturnExiledCardsWithCounter) so no owner
// scope, destination, or counter filter is silently dropped.
func TestParseReturnExiledCardsWithCounterFailsClosed(t *testing.T) {
	for _, source := range []string{
		// Opponent owner scope is a different, unsupported semantic.
		"Put all exiled cards an opponent owns with intel counters on them into your hand.",
		// A destination other than your hand.
		"Put all exiled cards you own with intel counters on them onto the battlefield.",
		// An unknown counter word is not a named marker counter.
		"Put all exiled cards you own with bogus counters on them into your hand.",
	} {
		document, _ := Parse(source, Context{InstantOrSorcery: true})
		if len(document.Abilities) == 1 && len(document.Abilities[0].Sentences) == 1 {
			for _, effect := range document.Abilities[0].Sentences[0].Effects {
				if effect.Kind == EffectReturnExiledCardsWithCounter {
					t.Fatalf("source %q was recognized as an exact return effect, want fail-closed", source)
				}
			}
		}
	}
}
