package cardgen

import (
	"github.com/natefinch/council4/cardgen/oracle/compiler"
	"github.com/natefinch/council4/cardgen/oracle/parser"
	"github.com/natefinch/council4/mtg/game"
)

// lowerDrainXLifeSpell handles the two-clause "drain" pattern
// "<target opponent | target player | each opponent | each player> loses X life
// and you gain X life", where the two clauses move the same amount X out of the
// drained player(s) and into the controller. The amount X may be fixed
// ("loses 2 life and you gain 2 life") or a single shared "where X is <dynamic>"
// definition that the compiler attaches to one of the two clauses (the other
// carrying the bare variable X). Fell Beast of Mordor is the anchor:
// "target opponent loses X life and you gain X life, where X is the number of
// +1/+1 counters on it." emits a LoseLife on the target opponent and a GainLife
// for the controller, both reading the same source counter count.
//
// It emits the lose clause followed by the gain clause sharing one Quantity and
// fails closed unless every guard holds, so richer drains (separate-sentence
// "that much"/"life lost this way" forms owned by lowerGroupLinkedLifeSpell and
// lowerLifeLostThisWayDrain, conditional or modal drains, and unparsed dynamic
// definitions) keep failing the round-trip.
func lowerDrainXLifeSpell(ctx contentCtx) (game.AbilityContent, bool) {
	if len(ctx.content.Effects) != 2 ||
		len(ctx.content.Conditions) != 0 ||
		len(ctx.content.Modes) != 0 ||
		len(abilityKeywordsExcludingSelectorPredicates(ctx.content)) != 0 ||
		ctx.optional {
		return game.AbilityContent{}, false
	}
	lose := &ctx.content.Effects[0]
	gain := &ctx.content.Effects[1]
	if lose.Kind != compiler.EffectLose ||
		gain.Kind != compiler.EffectGain ||
		gain.Context != parser.EffectContextController ||
		gain.Connection != parser.EffectConnectionAnd ||
		lose.Negated || gain.Negated ||
		!lose.Exact || !gain.Exact {
		return game.AbilityContent{}, false
	}
	amount, ok := drainSharedAmount(lose, gain, ctx.content.References)
	if !ok {
		return game.AbilityContent{}, false
	}
	losePrimitive, targets, ok := drainLosePrimitive(lose, ctx.content.Targets, amount)
	if !ok {
		return game.AbilityContent{}, false
	}
	return game.Mode{
		Targets: targets,
		Sequence: []game.Instruction{
			{Primitive: losePrimitive},
			{Primitive: game.GainLife{Player: game.ControllerReference(), Amount: amount}},
		},
	}.Ability(), true
}

// drainSharedAmount resolves the single Quantity X that both drain clauses move.
// It accepts equal fixed amounts on both clauses, or a single "where X is
// <dynamic>" definition on one clause with the bare variable X on the other.
func drainSharedAmount(
	lose, gain *compiler.CompiledEffect,
	references []compiler.CompiledReference,
) (game.Quantity, bool) {
	loseAmount := lose.Amount
	gainAmount := gain.Amount
	if loseAmount.Known && gainAmount.Known {
		if loseAmount.Value < 1 || loseAmount.Value != gainAmount.Value {
			return game.Quantity{}, false
		}
		return game.Fixed(loseAmount.Value), true
	}
	definition, bare := drainDefinitionAndBare(lose, gain)
	if definition == nil ||
		definition.Amount.DynamicForm != compiler.DynamicAmountWhereX ||
		!bare.Amount.VariableX ||
		bare.Amount.DynamicKind != compiler.DynamicAmountNone {
		return game.Quantity{}, false
	}
	if !drainDynamicReferencesSafe(references, definition.Amount.DynamicKind) {
		return game.Quantity{}, false
	}
	dynamic, ok := lowerDynamicAmount(definition.Amount, game.SourcePermanentReference())
	if !ok {
		return game.Quantity{}, false
	}
	return game.Dynamic(dynamic), true
}

// drainDefinitionAndBare splits the two clauses into the one carrying the
// dynamic "where X is" definition and the one carrying the bare variable X. It
// returns nil when both or neither clause carries a dynamic amount.
func drainDefinitionAndBare(lose, gain *compiler.CompiledEffect) (definition, bare *compiler.CompiledEffect) {
	loseDynamic := lose.Amount.DynamicKind != compiler.DynamicAmountNone
	gainDynamic := gain.Amount.DynamicKind != compiler.DynamicAmountNone
	switch {
	case loseDynamic && !gainDynamic:
		return lose, gain
	case gainDynamic && !loseDynamic:
		return gain, lose
	default:
		return nil, nil
	}
}

// drainDynamicReferencesSafe reports whether the ability's references permit
// reading the dynamic amount from the source permanent. A source-counter-count
// definition ("+1/+1 counters on it") is parser-pinned to the source, so its
// reference is safe regardless of how the compiler bound the pronoun. Any other
// object-reading dynamic is only safe when the lone reference binds to the
// source or the triggering permanent (the self-trigger creature).
func drainDynamicReferencesSafe(
	references []compiler.CompiledReference,
	dynamicKind compiler.DynamicAmountKind,
) bool {
	if len(references) == 0 {
		return true
	}
	if len(references) != 1 {
		return false
	}
	if dynamicKind == compiler.DynamicAmountSourceCounterCount {
		return true
	}
	switch references[0].Binding {
	case compiler.ReferenceBindingSource, compiler.ReferenceBindingEventPermanent:
		return true
	default:
		return false
	}
}

// drainLosePrimitive builds the life-loss primitive for the drained player(s)
// from the lose clause's subject: a player group for "each opponent"/"each
// player", or the single player target for "target opponent"/"target player".
func drainLosePrimitive(
	lose *compiler.CompiledEffect,
	abilityTargets []compiler.CompiledTarget,
	amount game.Quantity,
) (game.Primitive, []game.TargetSpec, bool) {
	switch lose.Context {
	case parser.EffectContextEachOpponent:
		if len(abilityTargets) != 0 {
			return nil, nil, false
		}
		return game.LoseLife{PlayerGroup: game.OpponentsReference(), Amount: amount}, nil, true
	case parser.EffectContextEachPlayer:
		if len(abilityTargets) != 0 {
			return nil, nil, false
		}
		return game.LoseLife{PlayerGroup: game.AllPlayersReference(), Amount: amount}, nil, true
	case parser.EffectContextTarget:
		if len(abilityTargets) != 1 {
			return nil, nil, false
		}
		spec, ok := playerTargetSpec(abilityTargets[0])
		if !ok {
			return nil, nil, false
		}
		return game.LoseLife{Player: game.TargetPlayerReference(0), Amount: amount}, []game.TargetSpec{spec}, true
	case parser.EffectContextReferencedObjectController:
		if len(abilityTargets) != 0 {
			return nil, nil, false
		}
		recipient, ok := eventReferencedObjectControllerPlayer(lose.References)
		if !ok {
			return nil, nil, false
		}
		return game.LoseLife{Player: recipient, Amount: amount}, nil, true
	default:
		return nil, nil, false
	}
}

// eventReferencedObjectControllerPlayer resolves a "that creature's
// controller"/"its controller" drained subject to the controller of the
// permanent that fired the trigger. The lone subject reference binds to the
// triggering event's permanent (the attacking or blocking creature) or its
// related permanent (the opposing combatant in a became-blocked trigger), so the
// drained player is that permanent's controller: Gloom Sower ("Whenever this
// creature becomes blocked by a creature, that creature's controller loses 2 life
// and you gain 2 life."), Revenge of Ravens, and MacCready, Lamplight Mayor.
//
// It fails closed for any reference set other than a sole event-bound permanent
// reference — a source-bound subject ("you"), a target-bound subject (owned by
// the inherited-target life path), or a multi-reference clause — so only the
// event-controller drain reaches this recipient.
func eventReferencedObjectControllerPlayer(references []compiler.CompiledReference) (game.PlayerReference, bool) {
	if len(references) != 1 {
		return game.PlayerReference{}, false
	}
	object, ok := lowerObjectReference(references[0], referenceLoweringContext{AllowEvent: true})
	if !ok {
		return game.PlayerReference{}, false
	}
	return game.ObjectControllerReference(object), true
}
