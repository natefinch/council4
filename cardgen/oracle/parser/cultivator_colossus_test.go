package parser

import "testing"

const cultivatorColossusOracle = "Trample\n" +
	"Cultivator Colossus's power and toughness are each equal to the number of lands you control.\n" +
	"When this creature enters, you may put a land card from your hand onto the battlefield tapped. If you do, draw a card and repeat this process."

func TestParseCultivatorColossusComposableProcess(t *testing.T) {
	t.Parallel()
	document, diagnostics := Parse(cultivatorColossusOracle, Context{CardName: "Cultivator Colossus"})
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	if len(document.Abilities) != 3 {
		t.Fatalf("abilities = %d, want keyword, CDA, and trigger", len(document.Abilities))
	}
	trigger := document.Abilities[2]
	if trigger.Kind != AbilityTriggered || len(trigger.Sentences) != 2 {
		t.Fatalf("trigger = %#v", trigger)
	}
	effects := trigger.Sentences[0].Effects
	if len(effects) != 1 || effects[0].Kind != EffectRepeatProcess {
		t.Fatalf("effects = %#v, want one repeat process", effects)
	}
	repeat := effects[0]
	if !repeat.Exact || !repeat.RepeatUntilFailure || len(repeat.RepeatBody) != 2 {
		t.Fatalf("repeat = %#v", repeat)
	}
	put, draw := repeat.RepeatBody[0], repeat.RepeatBody[1]
	if put.Kind != EffectPut || !put.Optional || !put.EntersTapped ||
		put.FromZone.String() != "Hand" || put.ToZone.String() != "Battlefield" {
		t.Fatalf("put kind=%v optional=%v tapped=%v from=%v to=%v selector=%#v",
			put.Kind, put.Optional, put.EntersTapped, put.FromZone, put.ToZone, put.Selection)
	}
	if draw.Kind != EffectDraw || !draw.Amount.Known || draw.Amount.Value != 1 {
		t.Fatalf("draw effect = %#v", draw)
	}
	if len(trigger.ConditionClauses) != 1 ||
		trigger.ConditionClauses[0].Predicate != ConditionPredicatePriorInstructionAccepted {
		t.Fatalf("conditions = %#v, want prior-instruction success", trigger.ConditionClauses)
	}
}
