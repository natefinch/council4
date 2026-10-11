package compiler

import (
	"testing"

	"github.com/natefinch/council4/cardgen/oracle/parser"
)

func TestActualCardQuantityProof(t *testing.T) {
	compilation, diagnostics := compileSource("Mill three cards. Draw a card. For each creature card milled this way, draw two cards.", pipelineContext{InstantOrSorcery: true})
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics: %#v", diagnostics)
	}
	effects := compilation.Abilities[0].Content.Effects
	amount := effects[2].Amount
	if !amount.ActualCardQuantitySupported() {
		t.Fatalf("quantity proof missing: %#v", amount)
	}
	if producer, ok := amount.ActualCardQuantityProducer(effects, effects[2].ClauseID); !ok || producer != 0 {
		t.Fatalf("producer=%d supported=%t", producer, ok)
	}
	for _, mutate := range []func(*CompiledAmount){
		func(a *CompiledAmount) { a.ProducerClauseID = 2 },
		func(a *CompiledAmount) { a.Multiplier++ },
		func(a *CompiledAmount) { a.actualCardResult = nil },
		func(a *CompiledAmount) { a.selector = &CompiledSelector{Kind: SelectorLand} },
	} {
		broken := amount
		mutate(&broken)
		if broken.ActualCardQuantitySupported() {
			t.Fatalf("mutated quantity retained proof: %#v", broken)
		}
	}
	for _, mutate := range []func([]CompiledEffect){
		func(e []CompiledEffect) { e[0].Kind = EffectDiscard },
		func(e []CompiledEffect) { e[0].Negated = true },
		func(e []CompiledEffect) { e[0].Exact = false },
		func(e []CompiledEffect) { e[0].ClauseID++ },
		func(e []CompiledEffect) { e[1].ClauseID = e[0].ClauseID },
		func(e []CompiledEffect) { e[2].DelayedTiming = 1 },
		func(e []CompiledEffect) { e[0].Replacement.Kind = parser.EffectReplacementInstead },
	} {
		broken := append([]CompiledEffect(nil), effects...)
		mutate(broken)
		if _, ok := amount.ActualCardQuantityProducer(broken, effects[2].ClauseID); ok {
			t.Fatal("mutated producer retained proof")
		}
		reindexed := append([]CompiledEffect{{ClauseID: 100, Kind: EffectDraw}}, effects...)
		if producer, ok := amount.ActualCardQuantityProducer(reindexed, effects[2].ClauseID); !ok || producer != 1 {
			t.Fatalf("typed producer lost after unrelated effect insertion: producer=%d ok=%t", producer, ok)
		}
	}
}
