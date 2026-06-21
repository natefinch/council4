package parser

import (
	"testing"
)

func TestParseRepeatProcessVariableX(t *testing.T) {
	t.Parallel()
	document, diagnostics := Parse(
		"Repeat the following process X times. Each opponent loses 3 life unless that player sacrifices a nonland permanent or discards a card.",
		Context{CardName: "Torment of Hailfire"})
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	var effects []EffectSyntax
	for i := range document.Abilities {
		for _, sentence := range document.Abilities[i].Sentences {
			effects = append(effects, sentence.Effects...)
		}
	}
	if len(effects) != 1 {
		t.Fatalf("effects = %#v, want exactly one", effects)
	}
	repeat := effects[0]
	if repeat.Kind != EffectRepeatProcess {
		t.Fatalf("kind = %v, want EffectRepeatProcess", repeat.Kind)
	}
	if !repeat.Amount.VariableX {
		t.Fatalf("amount = %+v, want VariableX", repeat.Amount)
	}
	if len(repeat.RepeatBody) != 1 {
		t.Fatalf("RepeatBody = %#v, want exactly one effect", repeat.RepeatBody)
	}
	if repeat.RepeatBody[0].Kind != EffectPunisherLoseLife {
		t.Fatalf("body kind = %v, want EffectPunisherLoseLife", repeat.RepeatBody[0].Kind)
	}
}

func TestParseRepeatProcessCardinalCount(t *testing.T) {
	t.Parallel()
	document, diagnostics := Parse(
		"Repeat the following process three times. Each opponent loses 2 life unless that player discards a card.",
		Context{CardName: "Test Repeat"})
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	var repeat *EffectSyntax
	for i := range document.Abilities {
		for si := range document.Abilities[i].Sentences {
			for ei := range document.Abilities[i].Sentences[si].Effects {
				if document.Abilities[i].Sentences[si].Effects[ei].Kind == EffectRepeatProcess {
					repeat = &document.Abilities[i].Sentences[si].Effects[ei]
				}
			}
		}
	}
	if repeat == nil {
		t.Fatal("no EffectRepeatProcess produced")
	}
	if repeat.Amount.VariableX || !repeat.Amount.Known || repeat.Amount.Value != 3 {
		t.Fatalf("amount = %+v, want fixed 3", repeat.Amount)
	}
}
