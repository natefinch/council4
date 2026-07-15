package compiler

import "testing"

// TestCompileSharedSubjectCombatTargetUnion verifies that the shared-subject
// three-way combat/target trigger union of Giggling Skitterspike compiles into
// three independent triggered abilities — one per event (attack, block, became
// the target of a spell) — that share the same self subject and the same
// source-power group-damage body. The parser distributes the shared "this
// creature" subject across the comma list, so downstream compilation sees three
// ordinary triggered abilities.
func TestCompileSharedSubjectCombatTargetUnion(t *testing.T) {
	t.Parallel()
	source := "Whenever this creature attacks, blocks, or becomes the target of a spell, it deals damage equal to its power to each opponent."
	compilation, diagnostics := compileSource(source, pipelineContext{})
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	if len(compilation.Abilities) != 3 {
		t.Fatalf("abilities = %d, want 3", len(compilation.Abilities))
	}

	type want struct {
		event       TriggerEvent
		stackObject TriggerStackObject
	}
	wants := []want{
		{event: TriggerEventAttackerDeclared},
		{event: TriggerEventBlockerDeclared},
		{event: TriggerEventObjectBecameTarget, stackObject: TriggerStackObjectSpell},
	}
	for i, ability := range compilation.Abilities {
		if ability.Trigger == nil {
			t.Fatalf("ability[%d] has no trigger", i)
		}
		if ability.Trigger.Pattern.Event != wants[i].event {
			t.Fatalf("ability[%d] event = %v, want %v", i, ability.Trigger.Pattern.Event, wants[i].event)
		}
		if ability.Trigger.Pattern.Source != TriggerSourceSelf {
			t.Fatalf("ability[%d] source = %v, want self", i, ability.Trigger.Pattern.Source)
		}
		if wants[i].event == TriggerEventObjectBecameTarget &&
			ability.Trigger.Pattern.StackObject != wants[i].stackObject {
			t.Fatalf("ability[%d] stack object = %v, want %v", i, ability.Trigger.Pattern.StackObject, wants[i].stackObject)
		}
		if len(ability.Content.Effects) != 1 ||
			ability.Content.Effects[0].Kind != EffectDealDamage {
			t.Fatalf("ability[%d] effects = %#v, want one deal-damage", i, ability.Content.Effects)
		}
		if ability.Content.Effects[0].Amount.DynamicKind != DynamicAmountSourcePower {
			t.Fatalf("ability[%d] amount = %#v, want source power", i, ability.Content.Effects[0].Amount)
		}
	}
}
