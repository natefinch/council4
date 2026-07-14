package compiler

import (
	"testing"

	"github.com/natefinch/council4/mtg/game/types"
)

func TestCompileSpelunkingSemantics(t *testing.T) {
	t.Parallel()
	const source = "When this enchantment enters, draw a card, then you may put a land card from your hand onto the battlefield. If you put a Cave onto the battlefield this way, you gain 4 life.\nLands you control enter untapped."
	compilation, diagnostics := compileSource(source, pipelineContext{CardName: "Spelunking"})
	if len(diagnostics) != 0 || len(compilation.Abilities) != 2 {
		t.Fatalf("compilation = %#v, diagnostics = %#v", compilation, diagnostics)
	}
	trigger := compilation.Abilities[0]
	if trigger.ExactSequence != ExactSequenceDrawPutLandSubtypeLife ||
		trigger.ExactSequencePutSubtype != types.Cave ||
		trigger.ExactSequenceLifeAmount != 4 {
		t.Fatalf("trigger = %#v", trigger)
	}
	if effects := compilation.Abilities[1].Content.Effects; len(effects) != 1 || !effects[0].EntersUntappedGroup() {
		t.Fatalf("static effects = %#v", effects)
	}
}
