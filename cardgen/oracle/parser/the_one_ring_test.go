package parser

import (
	"testing"

	"github.com/natefinch/council4/mtg/game/counter"
)

func TestParseTheOneRingDynamicAmounts(t *testing.T) {
	t.Parallel()
	document, diagnostics := Parse(
		"At the beginning of your upkeep, you lose 1 life for each burden counter on The One Ring.\n"+
			"{T}: Put a burden counter on The One Ring, then draw a card for each burden counter on The One Ring.",
		Context{CardName: "The One Ring"},
	)
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}

	var amounts []EffectAmountSyntax
	for _, ability := range document.Abilities {
		for _, sentence := range ability.Sentences {
			for _, effect := range sentence.Effects {
				if effect.Amount.DynamicKind != EffectDynamicAmountNone {
					amounts = append(amounts, effect.Amount)
				}
			}
		}
	}
	if len(amounts) != 2 {
		t.Fatalf("dynamic amounts = %#v, want life loss and draw", amounts)
	}
	for _, amount := range amounts {
		if amount.DynamicKind != EffectDynamicAmountSourceCounterCount ||
			amount.CounterKind != counter.Burden ||
			amount.ReferenceSpan.Start.Offset == 0 {
			t.Fatalf("amount = %#v, want burden counters on self", amount)
		}
	}
}

func TestParsePlayerProtectionUntilNextTurnIsExact(t *testing.T) {
	t.Parallel()
	document, diagnostics := Parse(
		"You gain protection from everything until your next turn.",
		Context{CardName: "Test Relic"},
	)
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	ability := document.Abilities[0]
	effect := ability.Sentences[0].Effects[0]
	if len(ability.SemanticKeywords) != 1 {
		t.Fatalf("keywords = %#v, want protection", ability.SemanticKeywords)
	}
	protection := ability.SemanticKeywords[0].Parameter.Protection()
	if !effect.Exact ||
		effect.Duration != EffectDurationUntilYourNextTurn ||
		effect.Context != EffectContextController ||
		ability.SemanticKeywords[0].Kind != KeywordProtection ||
		ability.SemanticKeywords[0].Parameter.Kind != KeywordParameterProtection ||
		!protection.Everything {
		t.Fatalf("effect = %#v, want exact temporary player protection", effect)
	}
}

func TestParsePlayerProtectionUnsupportedScopeIsInexact(t *testing.T) {
	t.Parallel()
	document, diagnostics := Parse(
		"You gain protection from red until your next turn.",
		Context{CardName: "Test Relic"},
	)
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	if effect := document.Abilities[0].Sentences[0].Effects[0]; effect.Exact {
		t.Fatalf("unsupported protection effect unexpectedly exact: %#v", effect)
	}
}
