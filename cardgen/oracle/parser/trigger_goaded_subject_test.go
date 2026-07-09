package parser

import "testing"

// TestParseGoadedSubjectAttackTrigger covers the "goaded" trigger-subject
// qualifier: "Whenever a goaded creature attacks, ..." parses to an attack
// trigger whose event subject selection records Goaded, so later layers can
// restrict the trigger to creatures that are goaded right now (Vengeful
// Ancestor's "Whenever a goaded creature attacks" ability).
func TestParseGoadedSubjectAttackTrigger(t *testing.T) {
	t.Parallel()
	document, diagnostics := Parse(
		"Whenever a goaded creature attacks, it deals 1 damage to its controller.",
		Context{},
	)
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	if len(document.Abilities) != 1 {
		t.Fatalf("got %d abilities, want 1", len(document.Abilities))
	}
	trigger := document.Abilities[0].Trigger
	if trigger == nil || trigger.TriggerEvent == nil {
		t.Fatalf("ability trigger = %#v, want a trigger-event clause", trigger)
	}
	if trigger.TriggerEvent.Kind != TriggerEventKindAttack {
		t.Fatalf("trigger kind = %v, want TriggerEventKindAttack", trigger.TriggerEvent.Kind)
	}
	if !trigger.TriggerEvent.Subject.Selection.Goaded {
		t.Fatal("trigger subject Selection.Goaded = false, want true (a goaded creature)")
	}
}

// TestParseUngoadedSubjectAttackTriggerOmitsGoaded confirms the qualifier is not
// spuriously set: a plain "Whenever a creature attacks" carries no Goaded flag,
// so the goaded restriction stays off unless the word "goaded" is present.
func TestParseUngoadedSubjectAttackTriggerOmitsGoaded(t *testing.T) {
	t.Parallel()
	document, diagnostics := Parse(
		"Whenever a creature attacks, it deals 1 damage to its controller.",
		Context{},
	)
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	trigger := document.Abilities[0].Trigger
	if trigger == nil || trigger.TriggerEvent == nil {
		t.Fatalf("ability trigger = %#v, want a trigger-event clause", trigger)
	}
	if trigger.TriggerEvent.Subject.Selection.Goaded {
		t.Fatal("trigger subject Selection.Goaded = true, want false (no \"goaded\" qualifier)")
	}
}

// TestConsumeTriggerSelectionModifiersGoaded unit-tests the modifier consumer
// directly so the "goaded" adjective maps onto the TriggerSelection.Goaded flag
// independent of the surrounding trigger grammar.
func TestConsumeTriggerSelectionModifiersGoaded(t *testing.T) {
	t.Parallel()
	var selection TriggerSelection
	remaining := consumeTriggerSelectionModifiers([]string{"goaded"}, &selection)
	if len(remaining) != 0 {
		t.Fatalf("remaining modifiers = %#v, want none consumed", remaining)
	}
	if !selection.Goaded {
		t.Fatal("selection.Goaded = false, want true")
	}
}
