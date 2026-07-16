package compiler

import (
	"testing"

	"github.com/natefinch/council4/cardgen/oracle/parser"
	"github.com/natefinch/council4/mtg/game"
)

const calamityGallopingInfernoOracle = "Haste\n" +
	"Whenever Calamity attacks while saddled, choose a nonlegendary creature that saddled it this turn and create a tapped and attacking token that's a copy of it. Sacrifice that token at the beginning of the next end step. Repeat this process once.\n" +
	"Saddle 1"

func TestCompileCalamityTypedProcess(t *testing.T) {
	t.Parallel()
	compilation, diagnostics := compileSource(calamityGallopingInfernoOracle, pipelineContext{CardName: "Calamity, Galloping Inferno"})
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	var repeat *CompiledEffect
	for i := range compilation.Abilities {
		for j := range compilation.Abilities[i].Content.Effects {
			effect := &compilation.Abilities[i].Content.Effects[j]
			if effect.Kind == EffectRepeatProcess {
				repeat = effect
			}
		}
	}
	if repeat == nil {
		t.Fatal("no compiled repeat process")
	}
	if !repeat.Exact || !repeat.Amount.Known || repeat.Amount.Value != 2 || len(repeat.RepeatBody) != 2 {
		t.Fatalf("repeat = %#v", repeat)
	}
	create := repeat.RepeatBody[0]
	if !create.Exact || !create.TokenCopyOfChosenSaddleContributor ||
		!create.TokenCopyEntersTapped || !create.TokenCopyAttacksWithSource {
		t.Fatalf("create = %#v", create)
	}
	sacrifice := repeat.RepeatBody[1]
	if !sacrifice.Exact || !sacrifice.CreatedTokensReference ||
		sacrifice.DelayedTiming != game.DelayedAtBeginningOfNextEndStep ||
		sacrifice.Context != parser.EffectContextController {
		t.Fatalf("sacrifice = %#v", sacrifice)
	}
}
