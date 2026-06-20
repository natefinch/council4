package parser

import (
	"testing"

	"github.com/natefinch/council4/mtg/game/zone"
)

func TestParseActivationCostReduction(t *testing.T) {
	t.Parallel()
	source := "Channel — {3}{U}, Discard this card: Return target artifact, creature, enchantment, or planeswalker to its owner's hand. This ability costs {1} less to activate for each legendary creature you control."
	document, diagnostics := Parse(source, Context{CardName: "Otawara, Soaring City"})
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	ability := document.Abilities[0]
	if len(ability.Sentences) != 2 || ability.Sentences[1].ActivationCostReduction == nil {
		t.Fatalf("sentences = %#v, want typed activation cost reduction", ability.Sentences)
	}
	reduction := ability.Sentences[1].ActivationCostReduction
	if reduction.PerObjectReduction != 1 ||
		reduction.Amount.DynamicKind != EffectDynamicAmountCount ||
		reduction.Amount.DynamicForm != EffectDynamicAmountFormForEach ||
		reduction.Amount.Multiplier != 1 ||
		reduction.Amount.Selection == nil ||
		reduction.Amount.Selection.Kind != SelectionCreature ||
		reduction.Amount.Selection.Controller != SelectionControllerYou ||
		reduction.Amount.Selection.Zone != zone.None {
		t.Fatalf("reduction = %#v", reduction)
	}
	if got := reduction.Amount.Selection.Supertypes; len(got) != 1 || got[0] != SupertypeLegendary {
		t.Fatalf("supertypes = %v, want legendary", got)
	}
	if !ability.Sentences[0].Effects[0].Exact || ability.Sentences[0].Effects[0].HasUnrecognizedSibling {
		t.Fatalf("bounce effect = %#v, want exact recognized sibling", ability.Sentences[0].Effects[0])
	}
	if coverage := DocumentCoverage(document); !coverage.Complete {
		t.Fatalf("coverage = %#v, want complete", coverage)
	}
}

func TestParseActivationCostReductionFailsClosed(t *testing.T) {
	t.Parallel()
	tests := []string{
		"Channel — {3}{U}, Discard this card: Draw a card. This spell costs {1} less to cast for each legendary creature you control.",
		"Channel — {3}{U}, Discard this card: Draw a card. This ability costs {X} less to activate for each legendary creature you control.",
		"Channel — {3}{U}, Discard this card: Draw a card. This ability costs {1} less to activate for each legendary creature card in your graveyard.",
		"Channel — {3}{U}, Discard this card: Draw a card. This ability costs {1} less to activate for legendary creatures you control.",
	}
	for _, source := range tests {
		t.Run(source, func(t *testing.T) {
			t.Parallel()
			document, _ := Parse(source, Context{CardName: "Test Channel"})
			if document.Abilities[0].Sentences[1].ActivationCostReduction != nil {
				t.Fatalf("source %q unexpectedly produced an activation cost reduction", source)
			}
		})
	}
}

func TestParseActivationCostReductionPreservesUnsupportedMainSentenceContent(t *testing.T) {
	t.Parallel()
	tests := []string{
		"{1}: Draw a card, then you become the monarch. This ability costs {1} less to activate for each legendary creature you control.",
		"{1}: Draw a card, then venture into the dungeon. This ability costs {1} less to activate for each legendary creature you control.",
	}
	for _, source := range tests {
		t.Run(source, func(t *testing.T) {
			t.Parallel()
			document, diagnostics := Parse(source, Context{CardName: "Test Card"})
			if len(diagnostics) != 0 {
				t.Fatalf("diagnostics = %#v", diagnostics)
			}
			ability := document.Abilities[0]
			if ability.Sentences[1].ActivationCostReduction == nil {
				t.Fatalf("source %q did not produce an activation cost reduction", source)
			}
			effects := ability.Sentences[0].Effects
			if len(effects) != 1 || effects[0].Exact {
				t.Fatalf("effects = %#v, want one inexact effect", effects)
			}
			if coverage := DocumentCoverage(document); coverage.Complete {
				t.Fatalf("coverage = %#v, want unsupported main-sentence content", coverage)
			}
		})
	}
}

func TestParseDiscardThisCardCostIsTypedSourceSelf(t *testing.T) {
	t.Parallel()
	document, diagnostics := Parse("Channel — {3}{U}, Discard this card: Draw a card.", Context{CardName: "Test Channel"})
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	component := document.Abilities[0].CostSyntax.Components[1]
	if component.Kind != CostComponentDiscard || !component.SourceSelf ||
		component.SourceZone != zone.Hand || !component.AmountKnown || component.AmountValue != 1 {
		t.Fatalf("discard component = %#v, want discard source card from hand", component)
	}
}
