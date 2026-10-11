package compiler

import (
	"slices"
	"testing"

	"github.com/natefinch/council4/cardgen/oracle/parser"
)

func TestControllerObservationOwnsExactTargetResource(t *testing.T) {
	for _, prefix := range []string{"", "Target player gains 1 life. ", "Exile target card from your graveyard. ", "Counter target spell. "} {
		content := compiledTargetActionBody(t, prefix+"You may look at the top card of target opponent's library. If you do, you gain 1 life.")
		effect := content.Effects[len(content.Effects)-2]
		if !content.ControllerObservesTargetLibraryTop(effect) {
			t.Fatalf("prefix %q lost canonical controller observation", prefix)
		}
	}
	first := compiledTargetActionBody(t, "You may look at the top card of target opponent's library.")
	second := compiledTargetActionBody(t, "You may look at the top card of target opponent's library.")
	if first.ControllerObservesTargetLibraryTop(second.Effects[0]) ||
		!second.ControllerObservesTargetLibraryTop(second.Effects[0]) {
		t.Fatal("copied syntax lost its own proof or borrowed another compiled body's scope")
	}
}

func TestControllerObservationRefusesDriftedOrPlayerActedSubjects(t *testing.T) {
	for _, test := range []struct {
		name   string
		mutate func(*AbilityContent, *CompiledEffect)
	}{
		{"no scope", func(_ *AbilityContent, effect *CompiledEffect) { effect.subjectScope = nil }},
		{"wrong clause", func(_ *AbilityContent, effect *CompiledEffect) { effect.ClauseID++ }},
		{"wrong order", func(_ *AbilityContent, effect *CompiledEffect) { effect.Order.Start++ }},
		{"inexact", func(_ *AbilityContent, effect *CompiledEffect) { effect.Exact = false }},
		{"negated", func(_ *AbilityContent, effect *CompiledEffect) { effect.Negated = true }},
		{"draw", func(_ *AbilityContent, effect *CompiledEffect) { effect.Kind = EffectDraw }},
		{"discard", func(_ *AbilityContent, effect *CompiledEffect) { effect.Kind = EffectDiscard }},
		{"sacrifice", func(_ *AbilityContent, effect *CompiledEffect) { effect.Kind = EffectSacrifice }},
		{"reveal", func(_ *AbilityContent, effect *CompiledEffect) { effect.Kind = EffectReveal }},
		{"foreign context", func(_ *AbilityContent, effect *CompiledEffect) { effect.Context = parser.EffectContextEventPlayer }},
		{"explicit player", func(_ *AbilityContent, effect *CompiledEffect) { effect.Player = parser.EffectPlayerTargetOwner }},
		{"wrong resource", func(_ *AbilityContent, effect *CompiledEffect) {
			effect.CardSource = parser.EffectCardSourcePriorInstructionResult
		}},
		{"two cards", func(_ *AbilityContent, effect *CompiledEffect) { effect.Amount.Value = 2 }},
		{"unknown amount", func(_ *AbilityContent, effect *CompiledEffect) { effect.Amount.Known = false }},
		{"X amount", func(_ *AbilityContent, effect *CompiledEffect) { effect.Amount.VariableX = true }},
		{"dynamic amount", func(_ *AbilityContent, effect *CompiledEffect) { effect.Amount.DynamicKind = DynamicAmountSourcePower }},
		{"foreign reference", func(_ *AbilityContent, effect *CompiledEffect) {
			effect.References = []CompiledReference{{Kind: ReferenceThatPlayer}}
		}},
		{"subject reference", func(_ *AbilityContent, effect *CompiledEffect) {
			effect.SubjectReferences = []CompiledReference{{Kind: ReferenceThatPlayer}}
		}},
		{"foreign target", func(content *AbilityContent, effect *CompiledEffect) {
			effect.Targets = compiledTargetActionBody(t, "Look at the top card of target opponent's library.").Effects[0].Targets
		}},
		{"plural target", func(_ *AbilityContent, effect *CompiledEffect) {
			effect.Targets = slices.Clone(effect.Targets)
			effect.Targets[0].Cardinality.Max = 2
		}},
		{"duplicate canonical clause", func(content *AbilityContent, effect *CompiledEffect) {
			owner := content.subjectScope.owner
			owner.Effects = append(owner.Effects, *effect)
		}},
		{"canonical resource drift", func(content *AbilityContent, _ *CompiledEffect) {
			content.subjectScope.owner.Effects[0].CardSource = parser.EffectCardSourcePriorInstructionResult
		}},
		{"canonical amount drift", func(content *AbilityContent, _ *CompiledEffect) {
			content.subjectScope.owner.Effects[0].Amount.Value = 2
		}},
	} {
		t.Run(test.name, func(t *testing.T) {
			content := compiledTargetActionBody(t, "You may look at the top card of target opponent's library. If you do, you gain 1 life.")
			effect := content.Effects[0]
			test.mutate(&content, &effect)
			if content.ControllerObservesTargetLibraryTop(effect) {
				t.Fatal("drifted or player-acted subject became a controller observation")
			}
		})
	}
}
