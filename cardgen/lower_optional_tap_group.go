package cardgen

import (
	"github.com/natefinch/council4/cardgen/oracle/compiler"
	"github.com/natefinch/council4/cardgen/oracle/parser"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/opt"
)

const optionalTapGroupCountKey = game.ResultKey("optional-tap-group-count")

// lowerOptionalTapGroupScaledConsequence lowers the reusable attack-triggered
// body "you may tap X untapped <group> you control. If you do, <this creature
// gets +X/+0 until end of turn> and <deals X damage to the player or
// planeswalker it's attacking>." (Myr Battlesphere).
//
// The parser folds the "may tap X ... you control" clause onto an optional
// EffectTap whose amount is the variable X and whose selector restricts the
// controller's untapped permanents; the affirmative "If you do" gate is a
// ConditionPredicatePriorInstructionAccepted clause; and every consequence
// effect scales by that same X. It lowers to a TapChosenGroup that publishes the
// number tapped as X, followed by one gated instruction per consequence effect,
// each reading X through DynamicAmountChosenNumber. Anchoring the tap group and
// pump on the ability's controller and source (not on the tap group) keeps the X
// damage flowing even when the source has left the battlefield, and the pump a
// harmless no-op then, matching the card's ruling.
//
// The body is text-blind: any card whose attack trigger is an optional
// controller-group tap gating a run of the supported scaled consequences lowers
// through the same path. Any shape the runtime sequence cannot represent fails
// closed, leaving the body unsupported rather than lowering a wrong sequence.
func lowerOptionalTapGroupScaledConsequence(
	_ string,
	ctx contentCtx,
	_ *parser.Ability,
) (game.AbilityContent, bool) {
	if ctx.optional ||
		len(ctx.content.Effects) < 2 ||
		len(ctx.content.Conditions) != 1 ||
		len(ctx.content.Keywords) != 0 ||
		len(ctx.content.Modes) != 0 ||
		len(ctx.content.Targets) != 0 {
		return game.AbilityContent{}, false
	}
	if !attackDefendingPlayerEvent(ctx.triggerEvent) {
		return game.AbilityContent{}, false
	}
	condition := ctx.content.Conditions[0]
	if condition.Kind != compiler.ConditionIf ||
		condition.Negated ||
		condition.Predicate != compiler.ConditionPredicatePriorInstructionAccepted {
		return game.AbilityContent{}, false
	}

	tapEffect := ctx.content.Effects[0]
	if tapEffect.Kind != compiler.EffectTap ||
		!tapEffect.Optional ||
		tapEffect.Negated ||
		tapEffect.DelayedTiming != 0 ||
		len(tapEffect.Targets) != 0 ||
		tapEffect.Context != parser.EffectContextController ||
		!tapEffect.Amount.VariableX ||
		tapEffect.Amount.Known ||
		tapEffect.Amount.DynamicKind != compiler.DynamicAmountNone ||
		tapEffect.Amount.Multiplier != 0 ||
		tapEffect.Amount.Addend != 0 {
		return game.AbilityContent{}, false
	}
	// The tap must select the controller's untapped permanents, so the tap group
	// is exactly the "untapped <group> you control" the ability restricts.
	if tapEffect.Selector.Controller != compiler.ControllerYou || !tapEffect.Selector.Untapped {
		return game.AbilityContent{}, false
	}
	selection, ok := SelectionForSelector(tapEffect.Selector)
	if !ok {
		return game.AbilityContent{}, false
	}

	paidCount := game.Dynamic(game.DynamicAmount{
		Kind:      game.DynamicAmountChosenNumber,
		ResultKey: optionalTapGroupCountKey,
	})
	gate := opt.Val(game.InstructionResultGate{
		Key:       optionalTapGroupCountKey,
		Succeeded: game.TriTrue,
	})

	sequence := []game.Instruction{
		{
			Primitive: game.TapChosenGroup{
				ChooseFrom:   game.PlayerControlledGroup(game.ControllerReference(), selection),
				PublishCount: optionalTapGroupCountKey,
				Prompt:       "Tap any number of the matching untapped permanents you control.",
			},
			PublishResult: optionalTapGroupCountKey,
		},
	}

	// The remaining effects are the "If you do" consequence; each must lower to a
	// supported scaled instruction reading the published tap count.
	for i := 1; i < len(ctx.content.Effects); i++ {
		instruction, ok := lowerTapScaledConsequenceEffect(ctx.content.Effects[i], paidCount, gate)
		if !ok {
			return game.AbilityContent{}, false
		}
		sequence = append(sequence, instruction)
	}

	return game.Mode{Sequence: sequence}.Ability(), true
}

// lowerTapScaledConsequenceEffect lowers one "If you do, ... X ..." consequence
// effect against the published tap count. It supports the self-pump ("gets +X/+0
// until end of turn") and attacked-defender damage ("deals X damage to the
// player or planeswalker it's attacking") consequences, gating each on the tap
// having tapped at least one permanent, and fails closed for any other effect.
func lowerTapScaledConsequenceEffect(
	effect compiler.CompiledEffect,
	paidCount game.Quantity,
	gate opt.V[game.InstructionResultGate],
) (game.Instruction, bool) {
	if effect.Negated || len(effect.Targets) != 0 || effect.DelayedTiming != 0 {
		return game.Instruction{}, false
	}
	switch effect.Kind {
	case compiler.EffectModifyPT:
		if effect.Context != parser.EffectContextSource ||
			!effect.Exact ||
			!effect.Amount.VariableX ||
			!effect.PowerDelta.VariableX ||
			effect.PowerDelta.Negative ||
			!effect.ToughnessDelta.Known ||
			effect.ToughnessDelta.Value != 0 ||
			effect.ToughnessDelta.Negative ||
			effect.ToughnessDelta.VariableX {
			return game.Instruction{}, false
		}
		duration, ok := temporaryContinuousDuration(effect.Duration)
		if !ok {
			return game.Instruction{}, false
		}
		return game.Instruction{
			Primitive: game.ModifyPT{
				Object:         game.SourcePermanentReference(),
				PowerDelta:     paidCount,
				ToughnessDelta: game.Fixed(0),
				Duration:       duration,
			},
			ResultGate: gate,
		}, true
	case compiler.EffectDealDamage:
		if !effect.Exact ||
			!effect.Amount.VariableX ||
			effect.DamageRecipient.Reference != parser.DamageRecipientReferenceAttackedDefender ||
			len(effect.DamageRecipient.GroupSelectors) != 0 {
			return game.Instruction{}, false
		}
		return game.Instruction{
			Primitive: game.Damage{
				Amount:       paidCount,
				Recipient:    game.AttackedDefenderDamageRecipient(),
				DamageSource: opt.Val(game.SourcePermanentReference()),
			},
			ResultGate: gate,
		}, true
	default:
		return game.Instruction{}, false
	}
}
