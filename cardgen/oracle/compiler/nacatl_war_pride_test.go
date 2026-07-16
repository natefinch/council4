package compiler

import (
	"testing"

	"github.com/natefinch/council4/cardgen/oracle/parser"
	"github.com/natefinch/council4/mtg/game"
)

const nacatlWarPrideOracle = "This creature must be blocked by exactly one creature if able.\n" +
	"Whenever this creature attacks, create X tokens that are copies of it and that are tapped and attacking, where X is the number of creatures defending player controls. Exile the tokens at the beginning of the next end step."

func TestCompileNacatlWarPrideTypedMechanics(t *testing.T) {
	t.Parallel()
	compilation, diagnostics := compileSource(nacatlWarPrideOracle, pipelineContext{CardName: "Nacatl War-Pride"})
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	if len(compilation.Abilities) != 2 {
		t.Fatalf("abilities = %d, want 2", len(compilation.Abilities))
	}
	static := compilation.Abilities[0]
	if static.Static == nil || len(static.Static.Declarations) != 1 ||
		static.Static.Declarations[0].Rule == nil ||
		static.Static.Declarations[0].Rule.Kind != StaticRuleMustBeBlockedByExactlyOne {
		t.Fatalf("static semantics = %#v", static.Static)
	}
	triggered := compilation.Abilities[1]
	if len(triggered.Content.Effects) != 2 {
		t.Fatalf("effects = %#v", triggered.Content.Effects)
	}
	create := triggered.Content.Effects[0]
	if !create.Exact || !create.TokenCopyOfSource || !create.TokenCopyEntersTapped ||
		!create.TokenCopyAttacksDefender ||
		create.Amount.Selector().Controller != ControllerDefendingPlayer {
		t.Fatalf("create = %#v", create)
	}
	exile := triggered.Content.Effects[1]
	if !exile.Exact || !exile.CreatedTokensReference ||
		exile.DelayedTiming != game.DelayedAtBeginningOfNextEndStep ||
		exile.Context != parser.EffectContextController {
		t.Fatalf("exile = %#v", exile)
	}
}
