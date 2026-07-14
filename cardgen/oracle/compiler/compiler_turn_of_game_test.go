package compiler

import "testing"

// TestCompileEntersTappedUnlessTurnOfGame confirms the Starting Town wording
// lowers to a negated "unless" replacement condition carrying the per-player
// turn-of-game predicate and its threshold, so the land enters tapped except on
// the controller's first three turns of the game.
func TestCompileEntersTappedUnlessTurnOfGame(t *testing.T) {
	t.Parallel()
	source := "This land enters tapped unless it's your first, second, or third turn of the game."
	compilation, diagnostics := compileSource(source, pipelineContext{})
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	ability := compilation.Abilities[0]
	if ability.Kind != AbilityReplacement {
		t.Fatalf("kind = %v, want AbilityReplacement", ability.Kind)
	}
	if len(ability.Content.Effects) != 1 || ability.Content.Effects[0].Kind != EffectEnterTapped {
		t.Fatalf("effects = %#v", ability.Content.Effects)
	}
	if len(ability.Content.Conditions) != 1 {
		t.Fatalf("conditions = %#v", ability.Content.Conditions)
	}
	condition := ability.Content.Conditions[0]
	if condition.Kind != ConditionUnless ||
		condition.Predicate != ConditionPredicateControllerTurnOfGameAtMost ||
		condition.Threshold != 3 ||
		!condition.Negated {
		t.Fatalf("condition = %#v, want negated unless turn-of-game threshold 3", condition)
	}
}
