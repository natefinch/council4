package compiler

import (
	"testing"

	"github.com/natefinch/council4/cardgen/oracle/parser"
)

func TestCompileCavesOfChaosConditionalImpulse(t *testing.T) {
	t.Parallel()
	compilation, diagnostics := compileSource(
		"Whenever this creature attacks, exile the top card of your library. "+
			"If you've completed a dungeon, you may play that card this turn without paying its mana cost. "+
			"Otherwise, you may play that card this turn.",
		pipelineContext{CardName: "Caves of Chaos Adventurer"},
	)
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	content := compilation.Abilities[0].Content
	if len(content.Effects) != 2 ||
		len(content.Conditions) != 1 ||
		len(content.Targets) != 0 {
		t.Fatalf("content = %#v", content)
	}
	for i := range content.Effects {
		effect := content.Effects[i]
		if effect.Kind != EffectImpulseExile ||
			!effect.Exact ||
			effect.Optional ||
			!effect.Amount.Known ||
			effect.Amount.Value != 1 ||
			effect.Duration != DurationThisTurn ||
			effect.ImpulseWithoutPayingManaCost != (i == 0) {
			t.Fatalf("effect %d = %#v", i, effect)
		}
	}
	if content.Effects[1].Connection != parser.EffectConnectionOtherwise {
		t.Fatalf("otherwise connection = %v", content.Effects[1].Connection)
	}
	if content.Conditions[0].Predicate != ConditionPredicateControllerCompletedADungeon {
		t.Fatalf("condition = %#v", content.Conditions[0])
	}
}
