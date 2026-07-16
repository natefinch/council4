package parser

import "testing"

const inscriptionOfAbundanceText = "Kicker {2}{G}\n" +
	"Choose one. If this spell was kicked, choose any number instead.\n" +
	"• Put two +1/+1 counters on target creature.\n" +
	"• Target player gains X life, where X is the greatest power among creatures they control.\n" +
	"• Target creature you control fights target creature you don't control."

func TestParseInscriptionOfAbundanceModalKicker(t *testing.T) {
	t.Parallel()
	document, diagnostics := Parse(inscriptionOfAbundanceText, Context{InstantOrSorcery: true})
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	if len(document.Abilities) != 2 || document.Abilities[1].Modal == nil {
		t.Fatalf("abilities = %#v", document.Abilities)
	}
	modal := document.Abilities[1].Modal
	if modal.MinModes != 1 || modal.MaxModes != 1 ||
		modal.ChoiceBonus.Condition != ModalChoiceBonusConditionSpellKicked ||
		!modal.ChoiceBonus.ReplaceRange ||
		modal.ChoiceBonus.MinModes != 0 || modal.ChoiceBonus.MaxModes != 3 ||
		len(modal.Options) != 3 {
		t.Fatalf("modal = %#v", modal)
	}
	life := modal.Options[1].Sentences[0].Effects[0]
	if life.Amount.DynamicKind != EffectDynamicAmountGreatestPower ||
		life.Amount.Selection == nil ||
		life.Amount.Selection.Controller != SelectionControllerThatPlayer {
		t.Fatalf("life amount = %#v", life.Amount)
	}
	if got := len(modal.Options[2].Sentences[0].Targets); got != 2 {
		t.Fatalf("fight target count = %d, want 2", got)
	}
}

func TestParseModalKickerGrammarFailsClosed(t *testing.T) {
	t.Parallel()
	for _, source := range []string{
		"Choose one. If this spell was kicked, choose any number twice instead.\n• Draw a card.\n• Gain 1 life.",
		"Choose one. If this spell was cast, choose any number instead.\n• Draw a card.\n• Gain 1 life.",
		"Choose one. If this spell was kicked, choose all instead.\n• Draw a card.\n• Gain 1 life.",
	} {
		document, _ := Parse(source, Context{InstantOrSorcery: true})
		if len(document.Abilities) == 1 && document.Abilities[0].Modal != nil &&
			document.Abilities[0].Modal.ChoiceBonus.Condition == ModalChoiceBonusConditionSpellKicked {
			t.Fatalf("unsupported variant recognized as kicked modal grammar: %q", source)
		}
	}
}
