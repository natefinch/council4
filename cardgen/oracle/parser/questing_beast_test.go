package parser

import "testing"

func TestParseQuestingBeastComposableClauses(t *testing.T) {
	t.Parallel()
	const source = "Vigilance, deathtouch, haste\n" +
		"Questing Beast can't be blocked by creatures with power 2 or less.\n" +
		"Combat damage that would be dealt by creatures you control can't be prevented.\n" +
		"Whenever Questing Beast deals combat damage to an opponent, it deals that much damage to target planeswalker that player controls."

	document, diagnostics := Parse(source, Context{CardName: "Questing Beast"})
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	if len(document.Abilities) != 4 {
		t.Fatalf("abilities = %d, want 4", len(document.Abilities))
	}

	blocker := document.Abilities[1].StaticDeclarations
	if len(blocker) != 1 ||
		blocker[0].Kind != StaticDeclarationRule ||
		len(blocker[0].Rule.Qualifiers) != 1 ||
		blocker[0].Rule.Qualifiers[0].Kind != StaticRuleQualifierBlockerPowerOrLess ||
		blocker[0].Rule.Qualifiers[0].Amount != 2 {
		t.Fatalf("blocker declaration = %#v", blocker)
	}

	prevention := document.Abilities[2].StaticDeclarations
	if len(prevention) != 1 ||
		prevention[0].Kind != StaticDeclarationCombatDamagePreventionProhibition ||
		prevention[0].PreventionSource.Kind != SelectionCreature ||
		prevention[0].PreventionSource.Controller != SelectionControllerYou {
		t.Fatalf("prevention declaration = %#v", prevention)
	}

	trigger := document.Abilities[3]
	if len(trigger.Sentences) != 1 || len(trigger.Sentences[0].Effects) != 1 {
		t.Fatalf("trigger sentences = %#v", trigger.Sentences)
	}
	effect := trigger.Sentences[0].Effects[0]
	if effect.Amount.DynamicKind != EffectDynamicAmountTriggeringCounterCount {
		t.Fatalf("damage amount kind = %v, want triggering-event anaphor", effect.Amount.DynamicKind)
	}
	if len(effect.Targets) != 1 ||
		effect.Targets[0].Selection.Kind != SelectionPlaneswalker ||
		effect.Targets[0].Selection.Controller != SelectionControllerThatPlayer {
		t.Fatalf("damage target = %#v", effect.Targets)
	}
}

func TestParseCombatDamagePreventionProhibitionFailsClosed(t *testing.T) {
	t.Parallel()
	for _, source := range []string{
		"Damage that would be dealt by creatures you control can't be prevented.",
		"Combat damage that would be dealt to creatures you control can't be prevented.",
		"Combat damage that would be dealt by creatures can't be prevented.",
		"Combat damage that would be dealt by creatures you control can't be prevented this turn.",
		"Combat damage that would be dealt by creatures you control during your turn can't be prevented.",
	} {
		document, _ := Parse(source, Context{})
		if len(document.Abilities) > 0 && len(document.Abilities[0].StaticDeclarations) != 0 {
			t.Errorf("Parse(%q) produced declarations %#v", source, document.Abilities[0].StaticDeclarations)
		}
	}
}
