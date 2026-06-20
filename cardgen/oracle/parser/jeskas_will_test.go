package parser

import (
	"testing"

	"github.com/natefinch/council4/mtg/game/zone"
)

const jeskasWillText = "Choose one. If you control a commander as you cast this spell, you may choose both instead.\n" +
	"• Add {R} for each card in target opponent's hand.\n" +
	"• Exile the top three cards of your library. You may play them this turn."

func TestParseJeskasWillTypedSyntax(t *testing.T) {
	t.Parallel()
	document, diagnostics := Parse(jeskasWillText, Context{InstantOrSorcery: true})
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	if len(document.Abilities) != 1 || document.Abilities[0].Modal == nil {
		t.Fatalf("abilities = %#v", document.Abilities)
	}
	modal := document.Abilities[0].Modal
	if !modal.ChoiceKnown || modal.MinModes != 1 || modal.MaxModes != 1 ||
		modal.ChoiceBonus.Condition != ModalChoiceBonusConditionControlsCommander ||
		modal.ChoiceBonus.AdditionalMaxModes != 1 ||
		len(modal.Options) != 2 {
		t.Fatalf("modal = %#v", modal)
	}
	manaEffect := modal.Options[0].Sentences[0].Effects[0]
	if manaEffect.Kind != EffectAddMana || !manaEffect.Exact ||
		manaEffect.Amount.DynamicKind != EffectDynamicAmountCount ||
		manaEffect.Amount.DynamicForm != EffectDynamicAmountFormForEach ||
		manaEffect.Amount.Multiplier != 1 ||
		manaEffect.Amount.Selection == nil ||
		manaEffect.Amount.Selection.Controller != SelectionControllerOpponent ||
		manaEffect.Amount.Selection.Zone != zone.Hand {
		t.Fatalf("mana effect = %#v", manaEffect)
	}
	impulseEffect := modal.Options[1].Sentences[0].Effects[0]
	if impulseEffect.Kind != EffectImpulseExile || !impulseEffect.Exact ||
		!impulseEffect.Amount.Known || impulseEffect.Amount.Value != 3 ||
		impulseEffect.Duration != EffectDurationThisTurn {
		t.Fatalf("impulse effect = %#v", impulseEffect)
	}
}

func TestParseImpulseExileOutsideCommanderModal(t *testing.T) {
	t.Parallel()
	document, diagnostics := Parse(
		"Exile the top three cards of your library. You may play them this turn.",
		Context{InstantOrSorcery: true},
	)
	if len(diagnostics) != 0 || len(document.Abilities) != 1 {
		t.Fatalf("diagnostics = %#v, abilities = %#v", diagnostics, document.Abilities)
	}
	ability := document.Abilities[0]
	if len(ability.Sentences) != 2 ||
		len(ability.Sentences[0].Effects) != 1 ||
		ability.Sentences[0].Effects[0].Kind != EffectImpulseExile ||
		len(ability.SemanticReferences) != 0 {
		t.Fatalf("ability = %#v", ability)
	}
}

func TestParseJeskasWillVariantsFailClosed(t *testing.T) {
	t.Parallel()
	variants := []string{
		"Choose one. If you own a commander as you cast this spell, you may choose both instead.\n• Add {R} for each card in target opponent's hand.\n• Exile the top three cards of your library. You may play them this turn.",
		"Choose one. If you control a commander as you cast this spell, you may choose both instead.\n• Add {R} for each card in target player's hand.\n• Exile the top three cards of your library. You may play them this turn.",
		"Choose one. If you control a commander as you cast this spell, you may choose both instead.\n• Add {R} for each card in target opponent's hand.\n• Exile the top three cards of your library. You may play them.",
		"Choose one. If you control a commander as you cast this spell, you may choose both instead.\n• Add {R} for each card in target opponent's hand.\n• Exile the top three cards of your library. You may play them until your next turn.",
	}
	for _, source := range variants {
		document, _ := Parse(source, Context{InstantOrSorcery: true})
		if len(document.Abilities) == 1 && document.Abilities[0].Modal != nil {
			modal := document.Abilities[0].Modal
			fullyRecognized := modal.ChoiceBonus.Condition == ModalChoiceBonusConditionControlsCommander &&
				len(modal.Options) == 2 &&
				len(modal.Options[0].Sentences) > 0 &&
				len(modal.Options[0].Sentences[0].Effects) == 1 &&
				modal.Options[0].Sentences[0].Effects[0].Exact &&
				len(modal.Options[1].Sentences) > 0 &&
				len(modal.Options[1].Sentences[0].Effects) == 1 &&
				modal.Options[1].Sentences[0].Effects[0].Kind == EffectImpulseExile
			if fullyRecognized {
				t.Fatalf("variant was fully recognized:\n%s", source)
			}
		}
	}
}
