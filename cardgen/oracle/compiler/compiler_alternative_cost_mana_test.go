package compiler

import (
	"testing"

	"github.com/natefinch/council4/mtg/game/types"
)

func compileManaAlternative(t *testing.T, source string) *CompiledAlternativeCost {
	t.Helper()
	compilation, diagnostics := compileSource(source, pipelineContext{InstantOrSorcery: true})
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	if len(compilation.Abilities) == 0 {
		t.Fatal("no abilities compiled")
	}
	alternative := compilation.Abilities[0].AlternativeCost
	if alternative == nil || alternative.Kind != AlternativeCostMana {
		t.Fatalf("alternative cost = %#v, want mana-only kind", alternative)
	}
	return alternative
}

func TestCompileZeroManaAlternativeCost(t *testing.T) {
	t.Parallel()
	alternative := compileManaAlternative(t,
		"You may pay {0} rather than pay this spell's mana cost.\nDraw a card.")
	if alternative.Condition != AlternativeCostConditionUnknown {
		t.Fatalf("condition = %#v, want unconditional", alternative.Condition)
	}
	if alternative.ManaCost.String() != "{0}" {
		t.Fatalf("mana cost = %q, want {0}", alternative.ManaCost.String())
	}
	if alternative.WithoutPayingManaCost {
		t.Fatal("{0} alternative was flagged as without-paying-mana-cost")
	}
}

func TestCompileOpponentGainedLifeManaAlternativeCost(t *testing.T) {
	t.Parallel()
	alternative := compileManaAlternative(t,
		"If an opponent gained life this turn, you may pay {B} rather than pay this spell's mana cost.\n"+
			"Target player loses 5 life and you gain 5 life.")
	if alternative.Condition != AlternativeCostConditionOpponentGainedLifeThisTurn {
		t.Fatalf("condition = %#v, want opponent-gained-life", alternative.Condition)
	}
	if alternative.ManaCost.String() != "{B}" {
		t.Fatalf("mana cost = %q, want {B}", alternative.ManaCost.String())
	}
}

func TestCompileCreaturesAttackingManaAlternativeCost(t *testing.T) {
	t.Parallel()
	alternative := compileManaAlternative(t,
		"If exactly one creature is attacking, you may pay {W} rather than pay this spell's mana cost.\n"+
			"Destroy target attacking creature without flying.")
	if alternative.Condition != AlternativeCostConditionCreaturesAttacking {
		t.Fatalf("condition = %#v, want creatures-attacking", alternative.Condition)
	}
	if alternative.ConditionCount != 1 || !alternative.ConditionExactly {
		t.Fatalf("count/exactly = %d/%t, want 1/true",
			alternative.ConditionCount, alternative.ConditionExactly)
	}
	if alternative.ManaCost.String() != "{W}" {
		t.Fatalf("mana cost = %q, want {W}", alternative.ManaCost.String())
	}
}

// TestCompilePermanentsOnBattlefieldManaAlternativeCost proves Blasphemous
// Edict's board-state gate compiles to the typed permanents-on-battlefield
// condition, carrying the threshold on ConditionCount and the counted permanent
// type on ConditionPermanentType.
func TestCompilePermanentsOnBattlefieldManaAlternativeCost(t *testing.T) {
	t.Parallel()
	alternative := compileManaAlternative(t,
		"You may pay {B} rather than pay this spell's mana cost if there are thirteen or more creatures on the battlefield.\n"+
			"Each player sacrifices thirteen creatures of their choice.")
	if alternative.Condition != AlternativeCostConditionPermanentsOnBattlefield {
		t.Fatalf("condition = %#v, want permanents-on-battlefield", alternative.Condition)
	}
	if alternative.ConditionCount != 13 {
		t.Fatalf("count = %d, want 13", alternative.ConditionCount)
	}
	if alternative.ConditionPermanentType != types.Creature {
		t.Fatalf("permanent type = %#v, want creature", alternative.ConditionPermanentType)
	}
	if alternative.ConditionExactly {
		t.Fatal("board-state gate must never be an exact-count comparison")
	}
	if alternative.ManaCost.String() != "{B}" {
		t.Fatalf("mana cost = %q, want {B}", alternative.ManaCost.String())
	}
}
