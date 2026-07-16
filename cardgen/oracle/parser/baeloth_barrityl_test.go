package parser

import (
	"testing"

	"github.com/natefinch/council4/mtg/game/types"
)

func TestParseBaelothBarritylMechanics(t *testing.T) {
	t.Parallel()
	source := "Creatures your opponents control with power less than Baeloth Barrityl's power are goaded. (They attack each combat if able and attack a player other than you if able.)\n" +
		"Whenever a goaded attacking or blocking creature dies, you create a Treasure token."
	document, diagnostics := Parse(source, Context{CardName: "Baeloth Barrityl, Entertainer"})
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}

	if len(document.Abilities) != 2 {
		t.Fatalf("abilities = %#v, want two", document.Abilities)
	}
	statics := document.Abilities[0].StaticDeclarations
	if len(statics) != 1 {
		t.Fatalf("static declarations = %#v, want one", statics)
	}
	declaration := statics[0]
	if declaration.Rule.Subject.Kind != StaticRuleSubjectOpponentControlledCreatures ||
		declaration.Rule.Operation.Kind != StaticRuleOperationGoaded ||
		!declaration.Subject.Group.PowerLessThanSource {
		t.Fatalf("static declaration = %#v", declaration)
	}
	trigger := document.Abilities[1].Trigger
	if trigger == nil || trigger.TriggerEvent == nil {
		t.Fatalf("trigger = %#v, want typed trigger event", trigger)
	}
	selection := trigger.TriggerEvent.Subject.Selection
	if !selection.Goaded || selection.CombatState != TriggerSelectionAttackingOrBlocking {
		t.Fatalf("trigger selection = %#v", selection)
	}
}

func TestParseGoadedAttackerControllerCreatesTreasure(t *testing.T) {
	t.Parallel()
	document, diagnostics := Parse(
		"Whenever a goaded creature attacks, its controller creates a Treasure token.",
		Context{},
	)
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	ability := document.Abilities[0]
	if ability.Trigger == nil || ability.Trigger.TriggerEvent == nil ||
		!ability.Trigger.TriggerEvent.Subject.Selection.Goaded {
		t.Fatalf("trigger = %#v", ability.Trigger)
	}
	effect := ability.Sentences[0].Effects[0]
	if !effect.Exact ||
		effect.Context != EffectContextReferencedObjectController ||
		len(effect.Selection.SubtypesAny) != 1 ||
		effect.Selection.SubtypesAny[0] != types.Treasure {
		t.Fatalf("effect = %#v", effect)
	}
}
