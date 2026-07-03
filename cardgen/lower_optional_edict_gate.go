package cardgen

import (
	"github.com/natefinch/council4/cardgen/oracle/compiler"
	"github.com/natefinch/council4/cardgen/oracle/parser"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/opt"
)

// nonControllerOptionalActionDeclinedKey wires the non-controller optional
// action instruction ("target opponent may sacrifice ...") to the failure-gated
// consequence that resolves when that player declines it ("If they don't, ...").
const nonControllerOptionalActionDeclinedKey = game.ResultKey("non-controller-optional-action")

// nonControllerEdictActionContext reports whether an effect context names a
// single non-controller player who performs the offered edict: the target
// opponent, the triggering event's player, or the defending player. The
// controller ("you") is excluded — a controller optional action is the
// affirmative optional-flow family the shared optional-flow planner owns.
func nonControllerEdictActionContext(context parser.EffectContextKind) bool {
	switch context {
	case parser.EffectContextTarget,
		parser.EffectContextReferencedPlayer,
		parser.EffectContextEventPlayer,
		parser.EffectContextDefendingPlayer:
		return true
	default:
		return false
	}
}

// lowerNonControllerOptionalEdictGate lowers the non-controller negative
// resolving gate the parser types as
// ConditionPredicatePriorInstructionNotAccepted after a non-controller optional
// edict offer:
//
//	At the beginning of your end step, target opponent may sacrifice two
//	nonland, nontoken permanents of their choice. If they don't, you draw two
//	cards. (Rakdos, Patron of Chaos)
//
// The offered edict is performed by a non-controller player (the target
// opponent, the event player), so its optional instruction names that player as
// its OptionalActor: the engine asks the sacrificing player — not the ability's
// controller — whether to perform it (CR 603.3b). The edict publishes whether it
// happened, and every consequence instruction is gated on it having been
// declined (TriFalse), the exact complement of the affirmative "if they do"
// resolving-success gate.
//
// The edict action and the consequence are lowered compositionally through the
// shared per-effect path (contextForEffect + lowerContent), so the family stays
// text-blind and fails closed on any edict or consequence the backend cannot
// already lower on its own. The edict action must be a single optional
// non-controller SacrificePermanents; the consequence may be one or several
// instructions but may not itself carry an optional or result-gated envelope.
func lowerNonControllerOptionalEdictGate(
	cardName string,
	ctx contentCtx,
	syntax *parser.Ability,
) (game.AbilityContent, bool) {
	if ctx.optional ||
		len(ctx.content.Effects) < 2 ||
		len(ctx.content.Conditions) != 1 ||
		len(ctx.content.Keywords) != 0 ||
		len(ctx.content.Modes) != 0 {
		return game.AbilityContent{}, false
	}
	if ctx.content.Conditions[0].Predicate != compiler.ConditionPredicatePriorInstructionNotAccepted {
		return game.AbilityContent{}, false
	}
	action := ctx.content.Effects[0]
	if !action.Optional ||
		action.Negated ||
		action.DelayedTiming != 0 ||
		action.Kind != compiler.EffectSacrifice ||
		!nonControllerEdictActionContext(action.Context) {
		return game.AbilityContent{}, false
	}
	if syntax == nil {
		return game.AbilityContent{}, false
	}
	clauseSyntaxes := splitEffectSyntaxes(syntax, ctx.content.Effects)

	actionInstr, actionTargets, ok := lowerEdictOfferInstruction(cardName, ctx, &action, &clauseSyntaxes[0])
	if !ok {
		return game.AbilityContent{}, false
	}

	sequence := make([]game.Instruction, 0, len(ctx.content.Effects))
	sequence = append(sequence, actionInstr)
	consequence, ok := lowerEdictDeclinedConsequence(cardName, ctx, clauseSyntaxes)
	if !ok {
		return game.AbilityContent{}, false
	}
	sequence = append(sequence, consequence...)
	return game.Mode{
		Targets:  actionTargets,
		Sequence: sequence,
	}.Ability(), true
}

// lowerEdictOfferInstruction lowers the offered edict mandatorily (through the
// shared sacrifice path) and re-establishes it as an optional instruction the
// sacrificing player decides, publishing whether it happened. It returns the
// instruction, the edicted player's target spec, and whether the edict is a
// single unconditional non-controller SacrificePermanents the backend lowered
// cleanly on its own.
func lowerEdictOfferInstruction(
	cardName string,
	ctx contentCtx,
	action *compiler.CompiledEffect,
	clause *parser.Ability,
) (game.Instruction, []game.TargetSpec, bool) {
	mandatoryAction := *action
	mandatoryAction.Optional = false
	actionCtx := contextForEffect(ctx, &mandatoryAction)
	actionCtx.content.Conditions = nil
	actionContent, actionDiag := lowerContent(cardName, actionCtx, clause)
	if actionDiag != nil ||
		actionContent.IsModal() ||
		len(actionContent.SharedTargets) != 0 ||
		len(actionContent.Modes) != 1 ||
		len(actionContent.Modes[0].Sequence) != 1 {
		return game.Instruction{}, nil, false
	}
	actionInstr := actionContent.Modes[0].Sequence[0]
	if actionInstr.Optional ||
		actionInstr.PublishResult != "" ||
		actionInstr.ResultGate.Exists ||
		actionInstr.OptionalActor.Exists {
		return game.Instruction{}, nil, false
	}
	sacrifice, ok := actionInstr.Primitive.(game.SacrificePermanents)
	if !ok || sacrifice.PlayerGroup.Kind != game.PlayerGroupReferenceNone {
		return game.Instruction{}, nil, false
	}
	actionInstr.Optional = true
	actionInstr.OptionalActor = opt.Val(sacrifice.Player)
	actionInstr.PublishResult = nonControllerOptionalActionDeclinedKey
	return actionInstr, actionContent.Modes[0].Targets, true
}

// lowerEdictDeclinedConsequence lowers each consequence effect and gates it on
// the edict having been declined (TriFalse). The consequence carries no targets
// of its own — the ability's lone target is the edicted player — and may not
// itself carry an optional or result-gated envelope; any consequence needing a
// target or an unsupported effect fails the whole gate closed.
func lowerEdictDeclinedConsequence(
	cardName string,
	ctx contentCtx,
	clauseSyntaxes []parser.Ability,
) ([]game.Instruction, bool) {
	conditionSpan := ctx.content.Conditions[0].Span
	var sequence []game.Instruction
	for i := 1; i < len(ctx.content.Effects); i++ {
		consequence := ctx.content.Effects[i]
		consequenceCtx := contextForEffect(ctx, &consequence)
		consequenceCtx.content.Conditions = nil
		// Drop the gate's antecedent reference ("they"/"that player" in "If they
		// don't, ...") that lives in the condition span, so the consequence
		// lowers as its own effect without a stray event-player reference.
		consequenceCtx.content.References = referencesOutsideSpan(
			consequenceCtx.content.References, conditionSpan)
		if len(consequenceCtx.content.Targets) != 0 {
			return nil, false
		}
		consequenceContent, diag := lowerContent(cardName, consequenceCtx, &clauseSyntaxes[i])
		if diag != nil ||
			consequenceContent.IsModal() ||
			len(consequenceContent.SharedTargets) != 0 ||
			len(consequenceContent.Modes) != 1 ||
			len(consequenceContent.Modes[0].Targets) != 0 ||
			len(consequenceContent.Modes[0].Sequence) == 0 {
			return nil, false
		}
		for _, instr := range consequenceContent.Modes[0].Sequence {
			if instr.Optional ||
				instr.PublishResult != "" ||
				instr.ResultGate.Exists {
				return nil, false
			}
			instr.ResultGate = opt.Val(game.InstructionResultGate{
				Key:       nonControllerOptionalActionDeclinedKey,
				Succeeded: game.TriFalse,
			})
			sequence = append(sequence, instr)
		}
	}
	return sequence, true
}
