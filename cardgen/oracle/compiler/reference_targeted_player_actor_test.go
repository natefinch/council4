package compiler

import (
	"slices"
	"testing"

	"github.com/natefinch/council4/cardgen/oracle/parser"
)

const targetedActorBody = "Any number of target players each mill two cards. If you're the monarch, each of those players mills ten cards instead."

func TestTargetedPlayerActorCanonicalBody(t *testing.T) {
	t.Parallel()
	for _, prefix := range []string{"", "Counter target spell. "} {
		t.Run(prefix, func(t *testing.T) {
			t.Parallel()
			content := compiledTargetActionBody(t, prefix+targetedActorBody)
			for _, effect := range content.Effects {
				if effect.Kind != EffectMill {
					continue
				}
				target, ok := content.TargetedPlayerActor(effect)
				if !ok || target.Selector.Kind != SelectorPlayer || target.Cardinality.Max <= 1 {
					t.Fatalf("canonical whole-set actor did not retain its owned target declaration: clause=%d refs=%+v", effect.ClauseID, effect.References)
				}
			}
		})
	}
	for _, text := range []string{
		"Target player draws a card. " + targetedActorBody,
		targetedActorBody + " Target player draws a card.",
		targetedActorBody + " Up to two target players each draw a card.",
	} {
		content := compiledTargetActionBody(t, text)
		for _, effect := range content.Effects {
			if effect.Kind == EffectMill {
				if _, ok := content.TargetedPlayerActor(effect); ok {
					t.Fatal("partial actor set borrowed another player target declaration")
				}
			}
			content := compiledTargetActionBody(t, "Exile target card from a graveyard. "+targetedActorBody)
			effect := content.Effects[len(content.Effects)-1]
			if _, ok := content.TargetedPlayerActor(effect); ok {
				t.Fatal("card-bound readback acquired an unproved player-group role")
			}
			reference := effect.References[0]
			if reference.Binding != ReferenceBindingPriorInstructionResult || reference.Subject.targetedPlayerActor {
				t.Fatal("unmodeled card/group-role ambiguity retagged the original prior-card binding")
			}
		}
	}
}

func TestTargetedPlayerActorRefusesDriftedOwnership(t *testing.T) {
	t.Parallel()
	for _, test := range []struct {
		name   string
		mutate func(*AbilityContent, *CompiledEffect)
	}{
		{"missing scope", func(body *AbilityContent, _ *CompiledEffect) { body.subjectScope = nil }},
		{"unowned effect", func(_ *AbilityContent, effect *CompiledEffect) { effect.subjectScope = nil }},
		{"wrong clause", func(_ *AbilityContent, effect *CompiledEffect) { effect.ClauseID += 10 }},
		{"wrong order", func(_ *AbilityContent, effect *CompiledEffect) { effect.Order.Start++ }},
		{"missing complete body", func(body *AbilityContent, _ *CompiledEffect) { body.Targets = nil }},
		{"duplicate declaration", func(body *AbilityContent, _ *CompiledEffect) { body.Targets = append(body.Targets, body.Targets[0]) }},
		{"foreign declaration", func(body *AbilityContent, _ *CompiledEffect) {
			body.Targets[0] = compiledTargetActionBody(t, targetedActorBody).Targets[0]
		}},
		{"changed count", func(body *AbilityContent, _ *CompiledEffect) { body.Targets[0].Cardinality.Max++ }},
		{"changed selector", func(body *AbilityContent, _ *CompiledEffect) { body.Targets[0].Selector.Kind = SelectorOpponent }},
		{"changed target order", func(body *AbilityContent, _ *CompiledEffect) { body.Targets[0].Order.Start++ }},
		{"inexact declaration", func(body *AbilityContent, _ *CompiledEffect) { body.Targets[0].Exact = false }},
		{"foreign reference", func(_ *AbilityContent, effect *CompiledEffect) {
			effect.References[0] = compiledTargetActionBody(t, targetedActorBody).Effects[1].References[0]
		}},
		{"wrong reference node", func(_ *AbilityContent, effect *CompiledEffect) { effect.References[0].NodeID++ }},
		{"wrong reference order", func(_ *AbilityContent, effect *CompiledEffect) { effect.References[0].Order.Start++ }},
		{"default reference", func(_ *AbilityContent, effect *CompiledEffect) { effect.References[0] = CompiledReference{} }},
		{"wrong reference noun", func(_ *AbilityContent, effect *CompiledEffect) {
			effect.References[0].SubjectNoun = parser.ObjectNounCard
		}},
		{"extra readback reference", func(_ *AbilityContent, effect *CompiledEffect) {
			effect.References = append(effect.References, effect.References[0])
		}},
		{"duplicate canonical effect", func(body *AbilityContent, effect *CompiledEffect) {
			owner := body.subjectScope.owner
			owner.Effects = append(slices.Clone(owner.Effects), *effect)
		}},
		{"duplicate canonical reference", func(body *AbilityContent, effect *CompiledEffect) {
			owner := body.subjectScope.owner
			owner.References = append(slices.Clone(owner.References), effect.References[0])
		}},
	} {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			body := compiledTargetActionBody(t, targetedActorBody)
			body.Targets = slices.Clone(body.Targets)
			effect := body.Effects[1]
			effect.References = slices.Clone(effect.References)
			test.mutate(&body, &effect)
			if _, ok := body.TargetedPlayerActor(effect); ok {
				t.Fatal("drifted, default, foreign or partial ownership was accepted")
			}
		})
	}
}
