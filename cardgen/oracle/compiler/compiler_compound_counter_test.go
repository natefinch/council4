package compiler

import (
	"testing"

	"github.com/natefinch/council4/mtg/game/counter"
)

// TestCompileCompoundCounterPlacement proves the parser's compound multi-kind
// counter clause ("Put two +1/+1 counters and a flying counter on target
// creature.") propagates through the compiler: the primary +1/+1 placement rides
// the shared CounterKind field and the flying placement is carried as a single
// CompiledCounterPlacement in AdditionalCounterPlacements.
func TestCompileCompoundCounterPlacement(t *testing.T) {
	t.Parallel()
	compilation, diagnostics := compileSource(
		"Put two +1/+1 counters and a flying counter on target creature.",
		pipelineContext{InstantOrSorcery: true},
	)
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	effects := compilation.Abilities[0].Content.Effects
	if len(effects) != 1 || effects[0].Kind != EffectPut {
		t.Fatalf("effects = %#v, want one put", effects)
	}
	if effects[0].CounterKind != counter.PlusOnePlusOne {
		t.Fatalf("primary counter kind = %v, want +1/+1", effects[0].CounterKind)
	}
	extra := effects[0].AdditionalCounterPlacements
	if len(extra) != 1 || extra[0].Kind != counter.Flying || extra[0].Amount != 1 {
		t.Fatalf("additional placements = %#v, want one {Flying 1}", extra)
	}
}

// TestCompileSingleCounterPlacementNoAdditional guards the single-kind placement:
// no AdditionalCounterPlacements are produced, so existing counter cards keep
// their prior compiled shape.
func TestCompileSingleCounterPlacementNoAdditional(t *testing.T) {
	t.Parallel()
	compilation, diagnostics := compileSource(
		"Put a +1/+1 counter on target creature.",
		pipelineContext{InstantOrSorcery: true},
	)
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	effects := compilation.Abilities[0].Content.Effects
	if len(effects) != 1 || len(effects[0].AdditionalCounterPlacements) != 0 {
		t.Fatalf("effects = %#v, want one put with no additional placements", effects)
	}
}
