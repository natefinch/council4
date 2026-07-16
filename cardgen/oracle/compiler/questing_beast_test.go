package compiler

import (
	"testing"

	"github.com/natefinch/council4/cardgen/oracle/parser"
	"github.com/natefinch/council4/mtg/game/types"
)

func TestCompileQuestingBeastPreventionAndDamageTarget(t *testing.T) {
	t.Parallel()
	const source = "Combat damage that would be dealt by creatures you control can't be prevented.\n" +
		"Whenever Questing Beast deals combat damage to an opponent, it deals that much damage to target planeswalker that player controls."

	compilation, diagnostics := compileSource(source, pipelineContext{CardName: "Questing Beast"})
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}

	if len(compilation.Abilities) != 2 {
		t.Fatalf("abilities = %d, want 2", len(compilation.Abilities))
	}
	declarations := compilation.Abilities[0].Static.Declarations
	if len(declarations) != 1 ||
		declarations[0].Kind != StaticDeclarationCombatDamagePreventionProhibition ||
		declarations[0].CombatDamagePreventionProhibition == nil {
		t.Fatalf("prevention declarations = %#v", declarations)
	}
	sourceSelector := declarations[0].CombatDamagePreventionProhibition.Source
	if sourceSelector.Kind != SelectorCreature ||
		sourceSelector.Controller != ControllerYou {
		t.Fatalf("prevention source = %#v", sourceSelector)
	}
	if requiredTypes := sourceSelector.RequiredTypesAny(); len(requiredTypes) > 0 &&
		(len(requiredTypes) != 1 || requiredTypes[0] != types.Creature) {
		t.Fatalf("prevention source types = %#v", requiredTypes)
	}

	trigger := compilation.Abilities[1]
	if trigger.Trigger == nil ||
		trigger.Trigger.Pattern.Event != TriggerEventDamageDealt ||
		trigger.Trigger.Pattern.CombatQualifier != TriggerCombatDamage ||
		trigger.Trigger.Pattern.Player != TriggerPlayerOpponent {
		t.Fatalf("trigger = %#v", trigger.Trigger)
	}
	if len(trigger.Content.Targets) != 1 ||
		trigger.Content.Targets[0].Selector.Kind != SelectorPlaneswalker ||
		trigger.Content.Targets[0].Selector.Controller != ControllerThatPlayer {
		t.Fatalf("targets = %#v", trigger.Content.Targets)
	}
	effects := trigger.Content.Effects
	if len(effects) != 1 || effects[0].Amount.DynamicKind != DynamicAmountTriggeringCounterCount {
		t.Fatalf("effects = %#v", effects)
	}
}

func TestRecognizeCombatDamagePreventionProhibitionFromTypedNode(t *testing.T) {
	t.Parallel()
	ability := CompiledAbility{Kind: AbilityStatic}
	statics := []parser.StaticDeclarationSyntax{{
		Kind: parser.StaticDeclarationCombatDamagePreventionProhibition,
		PreventionSource: parser.SelectionSyntax{
			Kind:       parser.SelectionCreature,
			Controller: parser.SelectionControllerYou,
		},
	}}
	declaration, ok := recognizeStaticCombatDamagePreventionProhibitionDeclaration(ability, statics)
	if !ok ||
		declaration.Kind != StaticDeclarationCombatDamagePreventionProhibition ||
		declaration.CombatDamagePreventionProhibition == nil ||
		declaration.CombatDamagePreventionProhibition.Source.Kind != SelectorCreature ||
		declaration.CombatDamagePreventionProhibition.Source.Controller != ControllerYou {
		t.Fatalf("declaration = %#v ok = %v", declaration, ok)
	}
}
