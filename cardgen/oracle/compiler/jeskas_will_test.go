package compiler

import (
	"testing"

	"github.com/natefinch/council4/cardgen/oracle/parser"
)

func TestCompileJeskasWillTypedModalBonus(t *testing.T) {
	t.Parallel()
	source := "Choose one. If you control a commander as you cast this spell, you may choose both instead.\n" +
		"• Add {R} for each card in target opponent's hand.\n" +
		"• Exile the top three cards of your library. You may play them this turn."
	compilation, diagnostics := compileSource(source, pipelineContext{InstantOrSorcery: true})
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	content := compilation.Abilities[0].Content
	modal := content.Modes[0].Modal
	if modal == nil ||
		modal.MinModes != 1 || modal.MaxModes != 1 ||
		modal.Bonus.Condition != ModeChoiceBonusConditionControlsCommander ||
		modal.Bonus.AdditionalMaxModes != 1 ||
		len(content.Modes) != 2 {
		t.Fatalf("content = %#v", content)
	}
	if content.Modes[0].Content.Effects[0].Kind != EffectAddMana ||
		content.Modes[0].Content.Effects[0].Amount.DynamicKind != DynamicAmountCount ||
		content.Modes[1].Content.Effects[0].Kind != EffectImpulseExile {
		t.Fatalf("modes = %#v", content.Modes)
	}
}

func TestCompileModeChoiceBonusIsTextBlind(t *testing.T) {
	t.Parallel()
	document := parser.Document{Abilities: []parser.Ability{{
		Kind: parser.AbilitySpell,
		Text: "unrelated metadata",
		Modal: &parser.Modal{
			MinModes:    1,
			MaxModes:    1,
			ChoiceKnown: true,
			ChoiceBonus: parser.ModalChoiceBonusSyntax{
				Condition:          parser.ModalChoiceBonusConditionControlsCommander,
				AdditionalMaxModes: 1,
			},
			Options: []parser.Mode{{Text: "unrelated mode"}},
		},
	}}}
	compilation, _ := Compile(document, Context{})
	bonus := compilation.Abilities[0].Content.Modes[0].Modal.Bonus
	if bonus.Condition != ModeChoiceBonusConditionControlsCommander || bonus.AdditionalMaxModes != 1 {
		t.Fatalf("bonus = %#v", bonus)
	}
}
