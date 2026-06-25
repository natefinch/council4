package cardgen

import (
	"github.com/natefinch/council4/cardgen/oracle/compiler"
	"github.com/natefinch/council4/cardgen/oracle/parser"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/opt"
)

// coinFlipResultKey is the result key under which the coin-flip RollDie
// instruction publishes its outcome. The win and lose branch instructions read
// it through a result gate.
const coinFlipResultKey = game.ResultKey("coin-flip-result")

// A coin flip is a fair two-sided random draw (CR 705). It is modeled as a
// RollDie with two sides: heads (2) is a win, tails (1) is a loss. The branch
// instructions are gated on the published value.
const (
	coinFlipTails = 1
	coinFlipHeads = 2
)

// lowerCoinFlipSequence lowers a recognized "Flip a coin." outcome
// ("Flip a coin. If you win the flip, <effect>." and/or its lose branch,
// Tavern Swindler). It emits a RollDie with two sides that publishes its result
// under coinFlipResultKey, then lowers each branch's effects through the
// standard content path and gates every branch instruction on the matching flip
// result: the win branch on heads, the lose branch on tails. Because the
// branches lower through lowerContent, any supported non-targeted branch effect
// composes (gain life, lose life, source damage, return/sacrifice this
// creature). It fails closed unless every effect belongs to a coin-flip branch
// and the body carries no content-level targets, conditions, modes, or
// references, and unless each branch lowers to a single non-targeted,
// ungated instruction sequence.
func lowerCoinFlipSequence(cardName string, ctx contentCtx, syntax *parser.Ability) (game.AbilityContent, bool) {
	if syntax == nil || syntax.CoinFlip == nil || ctx.optional {
		return game.AbilityContent{}, false
	}
	effects := ctx.content.Effects
	if len(effects) == 0 {
		return game.AbilityContent{}, false
	}
	for i := range effects {
		if effects[i].CoinFlipBranch == compiler.CoinFlipBranchNone {
			return game.AbilityContent{}, false
		}
	}
	if len(ctx.content.Targets) != 0 ||
		len(ctx.content.Conditions) != 0 ||
		len(ctx.content.Modes) != 0 ||
		len(ctx.content.References) != 0 {
		return game.AbilityContent{}, false
	}

	publisher := game.Instruction{Primitive: game.RollDie{Sides: 2}}
	// appendCoinFlipEffects appends the win branch then the lose branch, so each
	// branch is a contiguous run of effects sharing a CoinFlipBranch value.
	var branches []resultGatedBranch
	for start := 0; start < len(effects); {
		branch := effects[start].CoinFlipBranch
		end := start + 1
		for end < len(effects) && effects[end].CoinFlipBranch == branch {
			end++
		}
		gate, ok := coinFlipBranchResult(branch)
		if !ok {
			return game.AbilityContent{}, false
		}
		branchContent, ok := lowerCoinFlipBranch(cardName, ctx, syntax, effects[start:end])
		if !ok {
			return game.AbilityContent{}, false
		}
		branches = append(branches, resultGatedBranch{
			predicate: game.InstructionResultGate{
				AmountRange: opt.Val(game.IntRange{Min: gate, Max: gate}),
			},
			sequence: branchContent.Modes[0].Sequence,
		})
		start = end
	}
	sequence, ok := assembleResultGatedBranches(publisher, coinFlipResultKey, branches)
	if !ok {
		return game.AbilityContent{}, false
	}
	return game.Mode{Sequence: sequence}.Ability(), true
}

// lowerCoinFlipBranch lowers one coin-flip branch's effects through the standard
// content path. It copies the branch effects with their coin-flip marker cleared
// so the recursive lowerContent does not re-enter lowerCoinFlipSequence, and
// fails closed for a modal, targeted, or empty branch result.
func lowerCoinFlipBranch(
	cardName string,
	ctx contentCtx,
	syntax *parser.Ability,
	branchEffects []compiler.CompiledEffect,
) (game.AbilityContent, bool) {
	effects := make([]compiler.CompiledEffect, len(branchEffects))
	copy(effects, branchEffects)
	for i := range effects {
		effects[i].CoinFlipBranch = compiler.CoinFlipBranchNone
	}
	branchCtx := ctx
	branchCtx.content = compiler.AbilityContent{Effects: effects}
	// Clear the coin-flip marker on the syntax copy so the recursive lowering
	// treats the branch as ordinary content instead of re-entering (and failing
	// closed on) the coin-flip path.
	branchSyntax := *syntax
	branchSyntax.CoinFlip = nil
	content, diagnostic := lowerContent(cardName, branchCtx, &branchSyntax)
	if diagnostic != nil ||
		content.IsModal() ||
		len(content.Modes) != 1 ||
		len(content.SharedTargets) != 0 ||
		len(content.Modes[0].Targets) != 0 {
		return game.AbilityContent{}, false
	}
	return content, true
}

// coinFlipBranchResult returns the published flip value that gates a branch:
// heads for the win branch, tails for the lose branch.
func coinFlipBranchResult(branch compiler.CoinFlipBranch) (int, bool) {
	switch branch {
	case compiler.CoinFlipBranchWin:
		return coinFlipHeads, true
	case compiler.CoinFlipBranchLose:
		return coinFlipTails, true
	}
	return 0, false
}
