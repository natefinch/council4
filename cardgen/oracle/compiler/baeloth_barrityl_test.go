package compiler

import (
	"testing"

	"github.com/natefinch/council4/cardgen/oracle/parser"
	"github.com/natefinch/council4/mtg/game/types"
)

func TestCompileBaelothBarritylMechanics(t *testing.T) {
	t.Parallel()
	source := "Creatures your opponents control with power less than Baeloth Barrityl's power are goaded. (They attack each combat if able and attack a player other than you if able.)\n" +
		"Whenever a goaded attacking or blocking creature dies, you create a Treasure token."
	compilation, diagnostics := compileSource(source, pipelineContext{CardName: "Baeloth Barrityl, Entertainer"})
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}

	if len(compilation.Abilities) != 2 {
		t.Fatalf("abilities = %#v, want two", compilation.Abilities)
	}
	static := compilation.Abilities[0].Static
	if static == nil || len(static.Declarations) != 1 {
		t.Fatalf("static = %#v", static)
	}
	declaration := static.Declarations[0]
	if declaration.Rule == nil || declaration.Rule.Kind != StaticRuleGoaded ||
		declaration.Group.Domain != StaticGroupBattlefield ||
		declaration.Group.Selection.Controller != ControllerOpponent ||
		!declaration.Group.Selection.PowerLessThanSource {
		t.Fatalf("static declaration = %#v", declaration)
	}
	trigger := compilation.Abilities[1].Trigger
	if trigger == nil ||
		trigger.Pattern.Event != TriggerEventPermanentDied ||
		!trigger.Pattern.SubjectSelection.Goaded ||
		trigger.Pattern.SubjectSelection.CombatState != TriggerCombatStateAttackingOrBlocking ||
		len(trigger.Pattern.SubjectSelection.RequiredTypes) != 1 ||
		trigger.Pattern.SubjectSelection.RequiredTypes[0] != types.Creature {
		t.Fatalf("trigger = %#v", trigger)
	}
}

func TestCompileGoadedAttackerControllerCreatesTreasure(t *testing.T) {
	t.Parallel()
	compilation, diagnostics := compileSource(
		"Whenever a goaded creature attacks, its controller creates a Treasure token.",
		pipelineContext{},
	)
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	ability := compilation.Abilities[0]
	effect := ability.Content.Effects[0]
	if ability.Trigger == nil ||
		ability.Trigger.Pattern.Event != TriggerEventAttackerDeclared ||
		!ability.Trigger.Pattern.SubjectSelection.Goaded {
		t.Fatalf("trigger = %#v", ability.Trigger)
	}
	if effect.Kind != EffectCreate ||
		effect.Context != parser.EffectContextReferencedObjectController ||
		len(effect.References) != 1 ||
		effect.References[0].Binding != ReferenceBindingEventPermanent {
		t.Fatalf("effect = %#v", effect)
	}
}
