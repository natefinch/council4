package compiler

import (
	"testing"

	"github.com/natefinch/council4/cardgen/oracle/parser"
	"github.com/natefinch/council4/mtg/game/types"
)

func TestCompileNightmareShepherdIsPositionBlind(t *testing.T) {
	t.Parallel()
	const oracle = "Flying\nWhenever another nontoken creature you control dies, you may exile it. If you do, create a token that's a copy of that creature, except it's 1/1 and it's a Nightmare in addition to its other types."
	document, diagnostics := parser.Parse(oracle, parser.Context{CardName: "Nightmare Shepherd"})
	if len(diagnostics) != 0 {
		t.Fatalf("parse diagnostics = %#v", diagnostics)
	}
	document.Abilities[1].Text = "downstream must not inspect this text"

	compilation, diagnostics := Compile(document, Context{})
	if len(diagnostics) != 0 {
		t.Fatalf("compile diagnostics = %#v", diagnostics)
	}
	ability := compilation.Abilities[1]
	if ability.Trigger == nil ||
		ability.Trigger.Pattern.Event != TriggerEventPermanentDied ||
		ability.Trigger.Pattern.Controller != ControllerYou ||
		!ability.Trigger.Pattern.ExcludeSelf ||
		!ability.Trigger.Pattern.SubjectSelection.NonToken {
		t.Fatalf("trigger = %#v", ability.Trigger)
	}
	if len(ability.Content.Effects) != 2 {
		t.Fatalf("effects = %#v", ability.Content.Effects)
	}
	exile, create := ability.Content.Effects[0], ability.Content.Effects[1]
	if exile.Kind != EffectExile || !exile.Optional || len(exile.References) != 1 {
		t.Fatalf("exile kind=%v optional=%v references=%d", exile.Kind, exile.Optional, len(exile.References))
	}
	if exile.References[0].Binding != ReferenceBindingEventCard {
		t.Fatalf("exile reference binding = %v, want event card", exile.References[0].Binding)
	}
	if create.Kind != EffectCreate || !create.TokenCopyOfReference || len(create.References) == 0 {
		t.Fatalf("create kind=%v copyReference=%v references=%d",
			create.Kind, create.TokenCopyOfReference, len(create.References))
	}
	if create.References[0].Binding != ReferenceBindingPriorInstructionResult ||
		create.References[0].PriorInstruction != 0 {
		t.Fatalf("create reference = binding:%v prior:%d, want prior instruction 0",
			create.References[0].Binding, create.References[0].PriorInstruction)
	}
	if !create.TokenCopyOverridePTKnown ||
		create.TokenCopyOverridePower != 1 || create.TokenCopyOverrideToughness != 1 ||
		!create.TokenCopyOverrideAdditiveTypes ||
		len(create.TokenCopyOverrideSubtypes) != 1 ||
		create.TokenCopyOverrideSubtypes[0] != types.Nightmare {
		t.Fatalf("create override = known:%v %d/%d additive:%v subtypes:%v",
			create.TokenCopyOverridePTKnown,
			create.TokenCopyOverridePower,
			create.TokenCopyOverrideToughness,
			create.TokenCopyOverrideAdditiveTypes,
			create.TokenCopyOverrideSubtypes)
	}
}
