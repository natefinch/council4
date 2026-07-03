package cardgen

import (
	"github.com/natefinch/council4/cardgen/oracle/compiler"
	"github.com/natefinch/council4/cardgen/oracle/parser"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/opt"
)

// groupMayHaveActionKey wires a multiplayer "may have" causative offer
// instruction ("any player may have Browbeat deal 5 damage to them", "any
// opponent may have it deal 4 damage to them") to the result-gated consequence
// that resolves on whether at least one offered player accepted it ("If no one
// does, ..." / "If a player does, ...").
const groupMayHaveActionKey = game.ResultKey("group-may-have-action")

// groupMayHaveScope maps a multiplayer "may have" chooser scope, encoded by the
// parser on the "have" grant's context, to the player group offered the caused
// action: "any player" offers every player and "any opponent" offers every
// opponent. It fails closed on every other context so the single-chooser
// contexts (target opponent, defending player, controller) stay with
// lowerMayHaveActionGate.
func groupMayHaveScope(context parser.EffectContextKind) (game.PlayerGroupReference, bool) {
	switch context {
	case parser.EffectContextEachPlayer:
		return game.AllPlayersReference(), true
	case parser.EffectContextEachOpponent:
		return game.OpponentsReference(), true
	default:
		return game.PlayerGroupReference{}, false
	}
}

// lowerGroupMayHaveActionGate lowers the resolving gate that follows a
// multiplayer "may have" causative offer, which the parser types with a
// structural "have" grant (its context encoding the chooser scope), the caused
// "deal N damage to them" action, and a
// ConditionPredicatePriorInstruction{Accepted,NotAccepted} link:
//
//	Any player may have Browbeat deal 5 damage to them. If no one does, target
//	player draws three cards. (Browbeat)
//
//	Any player may have Book Burning deal 6 damage to them. If no one does,
//	target player mills six cards. (Book Burning)
//
//	When this creature enters, any opponent may have it deal 4 damage to them.
//	If a player does, sacrifice this creature. (Vexing Devil)
//
// Every player in the group (every player, or every opponent) is offered the
// source's damage in turn, and each accepting player is dealt it. The offer
// publishes whether at least one player accepted, and every consequence
// instruction is gated on that collective decision — Accepted TriTrue for the
// affirmative "if a player does" branch, TriFalse for the negative "if no one
// does" branch.
//
// The damage magnitude is read from the compiled caused action (a fixed amount
// dealt to the "them" group members), and the offer's damage source is left
// unset so the runtime resolves it to the resolving object's source — correct
// for both a spell (Browbeat) and an enters ability (Vexing Devil). The
// consequence effects are lowered compositionally through the shared per-effect
// path (contextForEffect + lowerContent), so the family stays text-blind and
// fails closed on any consequence the backend cannot already lower on its own.
// Unlike the single-chooser gate, the consequence may carry the ability's lone
// target (Browbeat's "target player draws", Book Burning's "target player
// mills"), which is collected onto the ability mode.
func lowerGroupMayHaveActionGate(
	cardName string,
	ctx contentCtx,
	syntax *parser.Ability,
) (game.AbilityContent, bool) {
	if ctx.optional ||
		len(ctx.content.Effects) < 3 ||
		len(ctx.content.Conditions) != 1 ||
		len(ctx.content.Keywords) != 0 ||
		len(ctx.content.Modes) != 0 {
		return game.AbilityContent{}, false
	}
	var accepted game.TriState
	switch ctx.content.Conditions[0].Predicate {
	case compiler.ConditionPredicatePriorInstructionAccepted:
		accepted = game.TriTrue
	case compiler.ConditionPredicatePriorInstructionNotAccepted:
		accepted = game.TriFalse
	default:
		return game.AbilityContent{}, false
	}
	have := ctx.content.Effects[0]
	action := ctx.content.Effects[1]
	if have.Kind != compiler.EffectGrantKeyword || !have.Optional || have.Negated {
		return game.AbilityContent{}, false
	}
	group, ok := groupMayHaveScope(have.Context)
	if !ok {
		return game.AbilityContent{}, false
	}
	if action.Kind != compiler.EffectDealDamage ||
		action.Negated ||
		action.DelayedTiming != 0 ||
		!action.Amount.Known ||
		action.Amount.Value <= 0 {
		return game.AbilityContent{}, false
	}
	if syntax == nil {
		return game.AbilityContent{}, false
	}

	offer := game.Instruction{
		Primitive: game.Damage{
			Amount:    game.Fixed(action.Amount.Value),
			Recipient: game.PlayerDamageRecipient(game.GroupOfferMemberReference()),
		},
		Optional:           true,
		OptionalActorGroup: opt.Val(group),
		PublishResult:      groupMayHaveActionKey,
	}

	clauseSyntaxes := splitEffectSyntaxes(syntax, ctx.content.Effects)
	sequence := make([]game.Instruction, 0, len(ctx.content.Effects))
	sequence = append(sequence, offer)
	consequence, targets, ok := lowerGroupMayHaveConsequence(cardName, ctx, accepted, clauseSyntaxes)
	if !ok {
		return game.AbilityContent{}, false
	}
	sequence = append(sequence, consequence...)
	return game.Mode{
		Targets:  targets,
		Sequence: sequence,
	}.Ability(), true
}

// lowerGroupMayHaveConsequence lowers each consequence effect, gates it on
// whether at least one player accepted the offer (accepted), and collects the
// consequence's lone player target, if any, onto the ability mode. A consequence
// may name one target (Browbeat's "target player draws", Book Burning's "target
// player mills") but may not itself carry an optional or result-gated envelope;
// any consequence with more than one target or an unsupported effect fails the
// whole gate closed.
func lowerGroupMayHaveConsequence(
	cardName string,
	ctx contentCtx,
	accepted game.TriState,
	clauseSyntaxes []parser.Ability,
) ([]game.Instruction, []game.TargetSpec, bool) {
	conditionSpan := ctx.content.Conditions[0].Span
	var sequence []game.Instruction
	var targets []game.TargetSpec
	for i := 2; i < len(ctx.content.Effects); i++ {
		consequence := ctx.content.Effects[i]
		consequenceCtx := contextForEffect(ctx, &consequence)
		consequenceCtx.content.Conditions = nil
		// Drop the gate's antecedent reference ("no one"/"a player" in "If no
		// one does, ...") that lives in the condition span, so the consequence
		// lowers as its own effect without a stray gate-subject reference.
		consequenceCtx.content.References = referencesOutsideSpan(
			consequenceCtx.content.References, conditionSpan)
		consequenceContent, diag := lowerContent(cardName, consequenceCtx, &clauseSyntaxes[i])
		if diag != nil ||
			consequenceContent.IsModal() ||
			len(consequenceContent.SharedTargets) != 0 ||
			len(consequenceContent.Modes) != 1 ||
			len(consequenceContent.Modes[0].Sequence) == 0 {
			return nil, nil, false
		}
		if modeTargets := consequenceContent.Modes[0].Targets; len(modeTargets) != 0 {
			// The consequence carries the ability's lone target (the player who
			// draws or mills). More than one target across the whole gate is
			// unsupported: the offer contributes none and the runtime binds the
			// consequence effect to target 0.
			if len(targets)+len(modeTargets) > 1 {
				return nil, nil, false
			}
			targets = append(targets, modeTargets...)
		}
		for _, instr := range consequenceContent.Modes[0].Sequence {
			if instr.Optional ||
				instr.PublishResult != "" ||
				instr.ResultGate.Exists {
				return nil, nil, false
			}
			instr.ResultGate = opt.Val(game.InstructionResultGate{
				Key:      groupMayHaveActionKey,
				Accepted: accepted,
			})
			sequence = append(sequence, instr)
		}
	}
	return sequence, targets, true
}
