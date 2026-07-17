package compiler

import (
	"testing"

	"github.com/natefinch/council4/cardgen/oracle/parser"
	"github.com/natefinch/council4/mtg/game"
)

const flamerushRiderOracle = "Whenever this creature attacks, create a token that's a copy of another target attacking creature and that's tapped and attacking. Exile the token at end of combat.\n" +
	"Dash {2}{R}{R} (You may cast this spell for its dash cost. If you do, it gains haste, and it's returned from the battlefield to its owner's hand at the beginning of the next end step.)"

func TestCompileFlamerushRiderTypedAttackCopy(t *testing.T) {
	t.Parallel()
	compilation, diagnostics := compileSource(flamerushRiderOracle, pipelineContext{CardName: "Flamerush Rider"})
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	effects := compilation.Abilities[0].Content.Effects
	if len(effects) != 2 {
		t.Fatalf("effects = %d, want create and delayed exile", len(effects))
	}
	create := effects[0]
	if !create.Exact || !create.TokenCopyOfTarget ||
		!create.TokenCopyEntersTapped || !create.TokenCopyAttacksWithTarget ||
		len(create.Targets) != 1 ||
		!create.Targets[0].Selector.Another ||
		!create.Targets[0].Selector.Attacking {
		t.Fatalf("create exact=%v copy=%v tapped=%v attacking=%v targets=%#v",
			create.Exact, create.TokenCopyOfTarget, create.TokenCopyEntersTapped,
			create.TokenCopyAttacksWithTarget, create.Targets)
	}
	exile := effects[1]
	if !exile.Exact || !exile.CreatedTokensReference ||
		exile.DelayedTiming != game.DelayedAtEndOfCombat ||
		exile.Context != parser.EffectContextController {
		t.Fatalf("exile exact=%v created=%v timing=%v context=%v targets=%#v refs=%#v text=%q",
			exile.Exact, exile.CreatedTokensReference, exile.DelayedTiming, exile.Context,
			exile.Targets, exile.References, exile.Text)
	}
}
