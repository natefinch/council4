package compiler

import (
	"testing"

	"github.com/natefinch/council4/cardgen/oracle/parser"
)

func TestCompileKickedModalReplacementRange(t *testing.T) {
	t.Parallel()
	source := "Choose one. If this spell was kicked, choose any number instead.\n" +
		"• Put two +1/+1 counters on target creature.\n" +
		"• Target player gains X life, where X is the greatest power among creatures they control.\n" +
		"• Target creature you control fights target creature you don't control."
	compilation, diagnostics := compileSource(source, pipelineContext{InstantOrSorcery: true})
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	content := compilation.Abilities[0].Content
	modal := content.Modes[0].Modal
	if modal == nil ||
		modal.MinModes != 1 || modal.MaxModes != 1 ||
		modal.Bonus.Condition != ModeChoiceBonusConditionSpellKicked ||
		!modal.Bonus.ReplaceRange ||
		modal.Bonus.MinModes != 0 || modal.Bonus.MaxModes != 3 ||
		len(content.Modes) != 3 {
		t.Fatalf("content = %#v", content)
	}
	if content.Modes[1].Content.Effects[0].Amount.DynamicKind != DynamicAmountGreatestPower ||
		content.Modes[1].Content.Effects[0].Amount.Selector().Controller != ControllerThatPlayer {
		t.Fatalf("life mode = %#v", content.Modes[1])
	}
	life := content.Modes[1].Content
	if len(life.References) != 1 ||
		life.References[0].Binding != ReferenceBindingTarget ||
		life.References[0].Pronoun != ReferencePronounThey ||
		life.References[0].Span != life.Effects[0].Amount.ReferenceSpan {
		t.Fatalf("life references = %#v, amount = %#v", life.References, life.Effects[0].Amount)
	}
}

func TestCompileKickedModalReplacementIsTextBlind(t *testing.T) {
	t.Parallel()
	document := parser.Document{Abilities: []parser.Ability{{
		Kind: parser.AbilitySpell,
		Text: "unrelated metadata",
		Modal: &parser.Modal{
			MinModes:    1,
			MaxModes:    1,
			ChoiceKnown: true,
			ChoiceBonus: parser.ModalChoiceBonusSyntax{
				Condition:    parser.ModalChoiceBonusConditionSpellKicked,
				ReplaceRange: true,
				MinModes:     0,
				MaxModes:     7,
			},
			Options: []parser.Mode{{Text: "unrelated mode"}},
		},
	}}}
	compilation, _ := Compile(document, Context{})
	bonus := compilation.Abilities[0].Content.Modes[0].Modal.Bonus
	if bonus.Condition != ModeChoiceBonusConditionSpellKicked ||
		!bonus.ReplaceRange || bonus.MinModes != 0 || bonus.MaxModes != 7 {
		t.Fatalf("bonus = %#v", bonus)
	}
}
