package cardgen

import (
	"github.com/natefinch/council4/cardgen/oracle/compiler"
	"github.com/natefinch/council4/cardgen/oracle/parser"
	"github.com/natefinch/council4/cardgen/oracle/shared"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/opt"
)

// mayHaveActionKey wires a "may have" causative offer instruction ("target
// opponent may have you draw ...", "you may have target player lose ...") to the
// result-gated consequence that resolves on whether the deciding player accepted
// it ("If they do, ..." / "If they don't, ...").
const mayHaveActionKey = game.ResultKey("may-have-action")

// mayHaveOfferPlan describes how a "may have" causative offer's deciding player
// is modeled from the "have" grant's context. The chooser is the player who
// decides whether the caused action happens; setOptionalActor reports whether
// that decision must be delegated with OptionalActor (a non-controller decides)
// or left to the controller (the runtime default, so no OptionalActor is set);
// chooserIsTarget reports whether the chooser is itself the ability's lone
// player target ("target opponent may have ...").
type mayHaveOfferPlan struct {
	chooser          game.PlayerReference
	setOptionalActor bool
	chooserIsTarget  bool
}

// mayHaveChooserPlan maps a "may have" chooser context to the plan that models
// its deciding player:
//
//   - target opponent ("target opponent may have you draw ...") both decides and
//     is the ability's lone player target, so OptionalActor is delegated to it.
//   - defending player ("defending player may have you draw ...") decides but is
//     a contextual player, not a target, so OptionalActor is delegated to it.
//   - controller ("you may have target player lose ...") decides as the runtime
//     default, so no OptionalActor is set and the caused action keeps its own
//     target (the recipient).
func mayHaveChooserPlan(context parser.EffectContextKind) (mayHaveOfferPlan, bool) {
	switch context {
	case parser.EffectContextTarget:
		return mayHaveOfferPlan{
			chooser:          game.TargetPlayerReference(0),
			setOptionalActor: true,
			chooserIsTarget:  true,
		}, true
	case parser.EffectContextDefendingPlayer:
		return mayHaveOfferPlan{
			chooser:          game.DefendingPlayerReference(),
			setOptionalActor: true,
			chooserIsTarget:  false,
		}, true
	case parser.EffectContextController:
		return mayHaveOfferPlan{
			chooser:          game.ControllerReference(),
			setOptionalActor: false,
			chooserIsTarget:  false,
		}, true
	default:
		return mayHaveOfferPlan{}, false
	}
}

// lowerMayHaveActionGate lowers the resolving gate that follows a "may have"
// causative offer, which the parser types with a structural "have" grant, the
// caused action, and a ConditionPredicatePriorInstruction{Accepted,NotAccepted}
// link:
//
//	When this creature enters, target opponent may have you create two Lander
//	tokens. If they don't, put two +1/+1 counters on this creature.
//	(Terrapact Intimidator)
//
//	Target opponent may have Risk Factor deal 4 damage to them. If that player
//	doesn't, you draw three cards. (Risk Factor)
//
//	Landfall — Whenever a land you control enters, you may have target player
//	lose 3 life. If you do, put three +1/+1 counters on Ob Nixilis.
//	(Ob Nixilis, the Fallen)
//
// A single player (the target opponent, the defending player, or the controller)
// decides whether the caused action happens. When a non-controller decides, the
// offer's optional instruction names that player as its OptionalActor even when
// the action's own actor is the controller ("you draw/create") or the source
// ("Risk Factor deal ... to them"); when the controller decides, the runtime
// default applies and no OptionalActor is set. The offer publishes whether it was
// accepted, and every consequence instruction is gated on that result — TriTrue
// for the affirmative "if they do" branch, TriFalse for the negative "if they
// don't" branch.
//
// The caused action and the consequence are lowered compositionally through the
// shared per-effect path (contextForEffect + lowerContent), so the family stays
// text-blind and fails closed on any action or consequence the backend cannot
// already lower on its own. The caused action must lower to a single unconditional
// non-gated instruction; the consequence may be one or several instructions but
// may not itself carry an optional or result-gated envelope.
func lowerMayHaveActionGate(
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
	var succeeded game.TriState
	switch ctx.content.Conditions[0].Predicate {
	case compiler.ConditionPredicatePriorInstructionAccepted:
		succeeded = game.TriTrue
	case compiler.ConditionPredicatePriorInstructionNotAccepted:
		succeeded = game.TriFalse
	default:
		return game.AbilityContent{}, false
	}
	have := ctx.content.Effects[0]
	action := ctx.content.Effects[1]
	if have.Kind != compiler.EffectGrantKeyword || !have.Optional || have.Negated {
		return game.AbilityContent{}, false
	}
	plan, ok := mayHaveChooserPlan(have.Context)
	if !ok {
		return game.AbilityContent{}, false
	}
	if action.Negated || action.DelayedTiming != 0 {
		return game.AbilityContent{}, false
	}
	if syntax == nil {
		return game.AbilityContent{}, false
	}
	clauseSyntaxes := splitEffectSyntaxes(syntax, ctx.content.Effects)

	offerInstr, offerTargets, ok := lowerMayHaveOfferInstruction(cardName, ctx, &action, plan, &clauseSyntaxes[1])
	if !ok {
		return game.AbilityContent{}, false
	}

	sequence := make([]game.Instruction, 0, len(ctx.content.Effects))
	sequence = append(sequence, offerInstr)
	consequence, ok := lowerMayHaveConsequence(cardName, ctx, succeeded, clauseSyntaxes)
	if !ok {
		return game.AbilityContent{}, false
	}
	sequence = append(sequence, consequence...)
	return game.Mode{
		Targets:  offerTargets,
		Sequence: sequence,
	}.Ability(), true
}

// lowerMayHaveOfferInstruction lowers the caused action mandatorily (through the
// shared per-effect path) and re-establishes it as an optional instruction the
// deciding player controls, publishing whether it was accepted. It returns the
// instruction, the ability's target spec, and whether the action is a single
// unconditional non-gated instruction the backend lowered cleanly on its own.
//
// When the chooser is the target opponent the ability names one player target:
// either the action already targets it (Risk Factor's damage) or the offer
// synthesizes it from the compiled target so OptionalActor's TargetPlayerReference
// resolves. When the chooser is the defending player the action must contribute no
// target — the defending player is the contextual chooser, not a target. When the
// controller decides, the caused action keeps its own target (the recipient, e.g.
// "target player" for Ob Nixilis) and no OptionalActor is set.
func lowerMayHaveOfferInstruction(
	cardName string,
	ctx contentCtx,
	action *compiler.CompiledEffect,
	plan mayHaveOfferPlan,
	clause *parser.Ability,
) (game.Instruction, []game.TargetSpec, bool) {
	mandatoryAction := *action
	mandatoryAction.Optional = false
	mandatoryAction.OptionalSpan = shared.Span{}
	mandatoryAction.RequiresOrderedLowering = false
	if !mandatoryAction.Exact && causativeActionForcibleExact(&mandatoryAction, ctx.content.Keywords) {
		// The caused action rides the base-verb form governed by "have" ("you
		// draw a card", "you create two Lander tokens"), so the parser leaves it
		// non-exact even though every runtime field is parsed. This is the same
		// structural-"have" artifact lowerOptionalHaveEffect clears; the single-
		// effect lowerer re-validates the parsed fields and fails closed on any
		// shape it cannot model.
		mandatoryAction.Exact = true
	}
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
	targets := actionContent.Modes[0].Targets
	switch {
	case plan.chooserIsTarget:
		// The target opponent both decides and (for a source-actor action)
		// receives the caused effect. If the action already bound the target
		// opponent it is target 0 and OptionalActor shares it; otherwise the
		// action is controller-scoped and the offer contributes the chooser as
		// the ability's sole target so OptionalActor's TargetPlayerReference(0)
		// resolves.
		if len(targets) == 0 {
			if len(ctx.content.Targets) != 1 {
				return game.Instruction{}, nil, false
			}
			spec, ok := playerTargetSpec(ctx.content.Targets[0])
			if !ok {
				return game.Instruction{}, nil, false
			}
			targets = []game.TargetSpec{spec}
		} else if len(targets) != 1 {
			return game.Instruction{}, nil, false
		}
	case !plan.setOptionalActor:
		// The controller decides, so the caused action keeps whatever target it
		// bound (the recipient, e.g. "target player lose 3 life") and the
		// decision rides the runtime default with no OptionalActor.
		if len(targets) > 1 {
			return game.Instruction{}, nil, false
		}
	default:
		// The defending player decides but is a contextual chooser, not a
		// target, so the caused action must contribute no target.
		if len(targets) != 0 {
			return game.Instruction{}, nil, false
		}
	}
	actionInstr.Optional = true
	if plan.setOptionalActor {
		actionInstr.OptionalActor = opt.Val(plan.chooser)
	}
	actionInstr.PublishResult = mayHaveActionKey
	return actionInstr, targets, true
}

// lowerMayHaveConsequence lowers each consequence effect and gates it on whether
// the offer was accepted (succeeded). The consequence carries no targets of its
// own — the ability's lone target, if any, is bound by the offer — and may not
// itself carry an optional or result-gated envelope; any consequence needing a
// target or an unsupported effect fails the whole gate closed.
func lowerMayHaveConsequence(
	cardName string,
	ctx contentCtx,
	succeeded game.TriState,
	clauseSyntaxes []parser.Ability,
) ([]game.Instruction, bool) {
	conditionSpan := ctx.content.Conditions[0].Span
	var sequence []game.Instruction
	for i := 2; i < len(ctx.content.Effects); i++ {
		consequence := ctx.content.Effects[i]
		consequenceCtx := contextForEffect(ctx, &consequence)
		consequenceCtx.content.Conditions = nil
		// Drop the gate's antecedent reference ("they"/"that player" in "If they
		// do/don't, ...") that lives in the condition span, so the consequence
		// lowers as its own effect without a stray chooser reference.
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
				Key:       mayHaveActionKey,
				Succeeded: succeeded,
			})
			sequence = append(sequence, instr)
		}
	}
	return sequence, true
}
