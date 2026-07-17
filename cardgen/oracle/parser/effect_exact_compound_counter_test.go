package parser

import (
	"testing"

	"github.com/natefinch/council4/mtg/game/counter"
)

// compoundCounterEffect parses a single counter-placement sentence and returns
// its resolving EffectPut, failing the test if the sentence does not shape into
// exactly one put effect.
func compoundCounterEffect(t *testing.T, source string) EffectSyntax {
	t.Helper()
	document, diagnostics := Parse(source, Context{InstantOrSorcery: true})
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

// TestParseCompoundCounterPlacementMultiKind proves the multi-kind compound
// counter clause Guide of Souls uses ("put two +1/+1 counters and a flying
// counter on target attacking creature.") parses with the primary +1/+1
// placement on CounterKind/CounterKnown and the flying placement carried as a
// single AdditionalCounterPlacements entry, and reconstructs exactly.
func TestParseCompoundCounterPlacementMultiKind(t *testing.T) {
	t.Parallel()
	effect := compoundCounterEffect(t,
		"Put two +1/+1 counters and a flying counter on target attacking creature.")
	if !effect.CounterKnown || effect.CounterKind != counter.PlusOnePlusOne {
		t.Fatalf("primary counter = (%v, known %t), want +1/+1 known", effect.CounterKind, effect.CounterKnown)
	}
	if len(effect.AdditionalCounterPlacements) != 1 {
		t.Fatalf("additional placements = %#v, want one", effect.AdditionalCounterPlacements)
	}
	extra := effect.AdditionalCounterPlacements[0]
	if extra.Kind != counter.Flying || extra.Amount != 1 {
		t.Fatalf("extra placement = %#v, want {Flying 1}", extra)
	}
	if !effect.Exact {
		t.Fatal("Exact = false, want true (compound clause must round-trip)")
	}
}

// TestParseCompoundCounterPlacementLifelink proves the compound clause is reused
// for other counter kinds ("Put a +1/+1 counter and a lifelink counter on target
// creature.", Unexpected Fangs): the primary +1/+1 rides the shared fields and
// the lifelink placement is the single additional entry.
func TestParseCompoundCounterPlacementLifelink(t *testing.T) {
	t.Parallel()
	effect := compoundCounterEffect(t,
		"Put a +1/+1 counter and a lifelink counter on target creature.")
	if !effect.CounterKnown || effect.CounterKind != counter.PlusOnePlusOne {
		t.Fatalf("primary counter = (%v, known %t), want +1/+1 known", effect.CounterKind, effect.CounterKnown)
	}
	if len(effect.AdditionalCounterPlacements) != 1 ||
		effect.AdditionalCounterPlacements[0].Kind != counter.Lifelink ||
		effect.AdditionalCounterPlacements[0].Amount != 1 {
		t.Fatalf("additional placements = %#v, want one {Lifelink 1}", effect.AdditionalCounterPlacements)
	}
	if !effect.Exact {
		t.Fatal("Exact = false, want true")
	}
}

// TestParseSingleCounterPlacementHasNoAdditional guards the common single-kind
// placement: it carries no AdditionalCounterPlacements so existing counter cards
// are unaffected by the compound-clause support.
func TestParseSingleCounterPlacementHasNoAdditional(t *testing.T) {
	t.Parallel()
	effect := compoundCounterEffect(t, "Put a +1/+1 counter on target creature.")
	if len(effect.AdditionalCounterPlacements) != 0 {
		t.Fatalf("additional placements = %#v, want none for single-kind placement", effect.AdditionalCounterPlacements)
	}
}

// TestParseCompoundCounterUnknownKindFailsClosed proves a compound clause whose
// additional segment names a counter kind without complete runtime semantics
// (a finality counter) does not round-trip to an exact, lowerable production.
func TestParseCompoundCounterUnknownKindFailsClosed(t *testing.T) {
	t.Parallel()
	effect := compoundCounterEffect(t,
		"Put a +1/+1 counter and a finality counter on target creature.")
	if effect.Exact {
		t.Fatal("Exact = true, want false (finality counter has no runtime semantics)")
	}
}

func TestParseConditionPrefixedDelayedReturnWithCompoundCounters(t *testing.T) {
	t.Parallel()
	const source = "{2}, {T}: Exile another target creature you control. You may return that card to the battlefield under its owner's control. If you don't, at the beginning of the next end step, return that card to the battlefield under its owner's control with a vigilance counter and a lifelink counter on it."
	document, diagnostics := Parse(source, Context{CardName: "Test Protector"})
	if len(diagnostics) != 0 {
		t.Fatalf("Parse diagnostics = %#v", diagnostics)
	}
	var effects []EffectSyntax
	for _, sentence := range document.Abilities[0].Sentences {
		effects = append(effects, sentence.Effects...)
	}
	if len(effects) != 3 {
		t.Fatalf("effects = %#v, want three", effects)
	}
	returnEffect := effects[2]
	if returnEffect.Kind != EffectReturn || returnEffect.DelayedTiming != DelayedTimingNextEndStep {
		t.Fatalf("fallback return = %#v, want next-end-step return", returnEffect)
	}
	if !returnEffect.CounterKnown || returnEffect.CounterKind != counter.Vigilance ||
		len(returnEffect.AdditionalCounterPlacements) != 1 ||
		returnEffect.AdditionalCounterPlacements[0].Kind != counter.Lifelink ||
		returnEffect.AdditionalCounterPlacements[0].Amount != 1 {
		t.Fatalf("fallback counters = %v/%#v, want vigilance and lifelink", returnEffect.CounterKind, returnEffect.AdditionalCounterPlacements)
	}
}
