package cardgen

import (
	"github.com/natefinch/council4/cardgen/oracle/compiler"
	"github.com/natefinch/council4/cardgen/oracle/parser"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/opt"
)

// voteResultKey is the result key under which the Vote instruction publishes its
// signed tally margin (votes for the first option minus votes for the second).
// The arm instructions read it through a result gate whose amount range encodes
// the margin sign each arm requires.
const voteResultKey = game.ResultKey("vote-result")

// lowerVoteSequence lowers a recognized "Starting with you, each player votes for
// <A> or <B>." vote (CR 701.32) and its majority-gated arms. It emits a Vote
// instruction that publishes its signed tally margin under voteResultKey, then
// lowers each arm's effects through the standard content path and gates every
// arm instruction on the margin sign the arm requires: a positive margin means
// the first option won, a negative margin the second, and a tie a zero margin.
// Because the arms lower through lowerContent, any supported non-targeted arm
// effect composes. It fails closed unless every effect belongs to a vote arm and
// the body carries no content-level targets, conditions, modes, or references,
// and unless each arm lowers to a single non-targeted, ungated instruction
// sequence.
func lowerVoteSequence(cardName string, ctx contentCtx, syntax *parser.Ability) (game.AbilityContent, bool) {
	if syntax == nil || syntax.Vote == nil || ctx.optional {
		return game.AbilityContent{}, false
	}
	if len(syntax.Vote.Options) != 2 {
		return game.AbilityContent{}, false
	}
	effects := ctx.content.Effects
	if len(effects) == 0 {
		return game.AbilityContent{}, false
	}
	for i := range effects {
		if !effects[i].VoteArm {
			return game.AbilityContent{}, false
		}
	}
	if len(ctx.content.Targets) != 0 ||
		len(ctx.content.Conditions) != 0 ||
		len(ctx.content.Modes) != 0 ||
		len(ctx.content.References) != 0 {
		return game.AbilityContent{}, false
	}

	sequence := []game.Instruction{{
		Primitive:     game.Vote{Options: append([]string(nil), syntax.Vote.Options...)},
		PublishResult: voteResultKey,
	}}
	// Each arm is a contiguous run of effects sharing the same option/tie gate.
	for start := 0; start < len(effects); {
		option := effects[start].VoteArmOption
		tie := effects[start].VoteArmTieInclusive
		end := start + 1
		for end < len(effects) &&
			effects[end].VoteArmOption == option &&
			effects[end].VoteArmTieInclusive == tie {
			end++
		}
		gate, ok := voteArmAmountRange(option, tie)
		if !ok {
			return game.AbilityContent{}, false
		}
		armContent, ok := lowerVoteArm(cardName, ctx, syntax, effects[start:end])
		if !ok {
			return game.AbilityContent{}, false
		}
		for k := range armContent.Modes[0].Sequence {
			instruction := armContent.Modes[0].Sequence[k]
			if instruction.ResultGate.Exists {
				return game.AbilityContent{}, false
			}
			instruction.ResultGate = opt.Val(game.InstructionResultGate{
				Key:         voteResultKey,
				AmountRange: opt.Val(gate),
			})
			sequence = append(sequence, instruction)
		}
		start = end
	}
	return game.Mode{Sequence: sequence}.Ability(), true
}

// lowerVoteArm lowers one vote arm's effects through the standard content path.
// It copies the arm effects with their vote marker cleared so the recursive
// lowerContent does not re-enter lowerVoteSequence, and fails closed for a modal,
// targeted, or empty arm result.
func lowerVoteArm(
	cardName string,
	ctx contentCtx,
	syntax *parser.Ability,
	armEffects []compiler.CompiledEffect,
) (game.AbilityContent, bool) {
	effects := make([]compiler.CompiledEffect, len(armEffects))
	copy(effects, armEffects)
	for i := range effects {
		effects[i].VoteArm = false
		effects[i].VoteArmOption = 0
		effects[i].VoteArmTieInclusive = false
	}
	armCtx := ctx
	armCtx.content = compiler.AbilityContent{Effects: effects}
	// Clear the vote marker on the syntax copy so the recursive lowering treats
	// the arm as ordinary content instead of re-entering (and failing closed on)
	// the vote path.
	armSyntax := *syntax
	armSyntax.Vote = nil
	content, diagnostic := lowerContent(cardName, armCtx, &armSyntax)
	if diagnostic != nil ||
		content.IsModal() ||
		len(content.Modes) != 1 ||
		len(content.SharedTargets) != 0 ||
		len(content.Modes[0].Targets) != 0 {
		return game.AbilityContent{}, false
	}
	return content, true
}

// voteArmAmountRange returns the published-margin range that gates a vote arm.
// The Vote instruction publishes the signed margin (votes for option 0 minus
// votes for option 1), so the first option wins with a strictly positive margin,
// the second with a strictly negative margin, and a tie at zero. A tie-inclusive
// arm extends its range to include the zero-margin tie.
func voteArmAmountRange(option int, tieInclusive bool) (game.IntRange, bool) {
	const bound = game.NumPlayers
	switch option {
	case 0:
		if tieInclusive {
			return game.IntRange{Min: 0, Max: bound}, true
		}
		return game.IntRange{Min: 1, Max: bound}, true
	case 1:
		if tieInclusive {
			return game.IntRange{Min: -bound, Max: 0}, true
		}
		return game.IntRange{Min: -bound, Max: -1}, true
	}
	return game.IntRange{}, false
}
