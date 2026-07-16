package compiler

import (
	"testing"

	"github.com/natefinch/council4/cardgen/oracle/parser"
)

// TestCompileKumenasAwakeningUpkeepDraws proves the upkeep body compiles into a
// triggered ability whose two draw effects — an each-player base draw and a
// controller "instead" replacement — are gated by a city's-blessing condition.
// This is the reusable shape behind Kumena's Awakening's conditional draw and
// carries no card-name logic.
func TestCompileKumenasAwakeningUpkeepDraws(t *testing.T) {
	t.Parallel()
	compilation, diagnostics := compileSource(
		"At the beginning of your upkeep, each player draws a card. "+
			"If you have the city's blessing, instead only you draw a card.",
		pipelineContext{},
	)
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	if len(compilation.Abilities) != 1 {
		t.Fatalf("abilities = %d, want 1", len(compilation.Abilities))
	}
	ability := compilation.Abilities[0]
	if ability.Trigger == nil {
		t.Fatal("upkeep body did not compile as a triggered ability")
	}

	effects := ability.Content.Effects
	if len(effects) != 2 {
		t.Fatalf("effects = %d, want 2: %#v", len(effects), effects)
	}

	base := effects[0]
	if base.Kind != EffectDraw || base.Context != parser.EffectContextEachPlayer || !base.Exact {
		t.Fatalf("base effect = %+v, want exact each-player draw", base)
	}
	if base.Replacement.Kind != parser.EffectReplacementNone {
		t.Fatalf("base replacement = %v, want none", base.Replacement.Kind)
	}

	replacement := effects[1]
	if replacement.Kind != EffectDraw || replacement.Context != parser.EffectContextController || !replacement.Exact {
		t.Fatalf("replacement effect = %+v, want exact controller draw", replacement)
	}
	if replacement.Replacement.Kind != parser.EffectReplacementInstead {
		t.Fatalf("replacement kind = %v, want instead", replacement.Replacement.Kind)
	}

	found := false
	for _, condition := range ability.Content.Conditions {
		if condition.Predicate == ConditionPredicateControllerHasCityBlessing {
			found = true
		}
	}
	if !found {
		t.Fatalf("city's-blessing condition not compiled: %#v", ability.Content.Conditions)
	}
}
