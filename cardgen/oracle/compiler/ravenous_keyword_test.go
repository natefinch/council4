package compiler

import (
	"slices"
	"testing"

	"github.com/natefinch/council4/cardgen/oracle/parser"
	"github.com/natefinch/council4/mtg/game/types"
)

func TestCompileRavenousKeyword(t *testing.T) {
	t.Parallel()
	source := "Ravenous (This creature enters with X +1/+1 counters on it. If X is 5 or more, draw a card when it enters.)"
	compilation, diagnostics := compileSource(source, pipelineContext{CardName: "Test Ravenous"})
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	if len(compilation.Abilities) != 1 || len(compilation.Abilities[0].Content.Keywords) != 1 {
		t.Fatalf("abilities = %#v; want one keyword ability", compilation.Abilities)
	}
	keyword := compilation.Abilities[0].Content.Keywords[0]
	if keyword.Kind != parser.KeywordRavenous || keyword.ParameterKind != parser.KeywordParameterNone {
		t.Fatalf("keyword = %#v; want Ravenous with no parameter", keyword)
	}
}

func TestCompileDamagedPlayerControlledArtifactOrEnchantmentTarget(t *testing.T) {
	t.Parallel()
	source := "Whenever this creature deals combat damage to a player, destroy target artifact or enchantment that player controls."
	compilation, diagnostics := compileSource(source, pipelineContext{CardName: "Test Hammer"})
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	ability := compilation.Abilities[0]
	if ability.Trigger == nil || ability.Trigger.Pattern.Event != TriggerEventDamageDealt ||
		ability.Trigger.Pattern.DamageRecipient != TriggerDamageRecipientPlayer {
		t.Fatalf("trigger = %#v; want damage-to-player trigger", ability.Trigger)
	}
	if len(ability.Content.Targets) != 1 {
		t.Fatalf("targets = %#v; want one", ability.Content.Targets)
	}
	if len(ability.Content.Effects) != 1 ||
		ability.Content.Effects[0].Kind != EffectDestroy ||
		!ability.Content.Effects[0].Exact ||
		ability.Content.Effects[0].Context != parser.EffectContextController {
		effect := ability.Content.Effects[0]
		t.Fatalf("effect kind=%v exact=%v context=%v targets=%#v references=%#v",
			effect.Kind, effect.Exact, effect.Context, effect.Targets, ability.Content.References)
	}
	target := ability.Content.Targets[0]
	if target.Selector.Controller != ControllerThatPlayer ||
		!slices.Equal(target.Selector.RequiredTypesAny(), []types.Card{types.Artifact, types.Enchantment}) {
		t.Fatalf("target selector = %#v; want artifact-or-enchantment controlled by that player", target.Selector)
	}
	if len(ability.Content.References) != 2 ||
		ability.Content.References[1].Binding != ReferenceBindingEventPlayer {
		t.Fatalf("references = %#v; want source and event-player bindings", ability.Content.References)
	}
}
