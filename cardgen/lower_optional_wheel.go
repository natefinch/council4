package cardgen

import (
	"github.com/natefinch/council4/cardgen/oracle/compiler"
	"github.com/natefinch/council4/cardgen/oracle/parser"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/opt"
)

// lowerOptionalWheelDiscardDraw lowers the optional whole-hand "wheel" body
// "You may discard {your hand | all the cards in your hand}. If you do, draw
// that many cards." into an optional whole-hand Discard followed by a Draw that
// reads the published discard count and applies only when the controller chose
// to discard. The discard publishes the number of cards discarded under a
// result key; the draw's dynamic amount reads that key and its result gate
// requires the discard to have been accepted, so the controller draws exactly
// as many cards as they discarded. Lowering reads only typed compiled fields:
// an optional whole-hand discard, a "that many" dynamic draw, and a
// prior-instruction-accepted "if you do" condition.
func lowerOptionalWheelDiscardDraw(ctx contentCtx) (game.AbilityContent, bool) {
	c := ctx.content
	if ctx.optional ||
		len(c.Modes) != 0 ||
		len(c.Targets) != 0 ||
		len(c.Keywords) != 0 ||
		len(c.References) != 0 ||
		len(c.Effects) != 2 ||
		len(c.Conditions) != 1 {
		return game.AbilityContent{}, false
	}

	discard := c.Effects[0]
	if discard.Kind != compiler.EffectDiscard ||
		!discard.DiscardEntireHand ||
		!discard.Optional ||
		discard.Negated ||
		discard.DelayedTiming != 0 ||
		discard.Context != parser.EffectContextController ||
		len(discard.Targets) != 0 {
		return game.AbilityContent{}, false
	}

	draw := c.Effects[1]
	if draw.Kind != compiler.EffectDraw ||
		draw.Optional ||
		draw.Negated ||
		draw.DelayedTiming != 0 ||
		draw.Context != parser.EffectContextController ||
		draw.Amount.Known ||
		draw.Amount.DynamicKind != compiler.DynamicAmountTriggeringCounterCount ||
		len(draw.Targets) != 0 {
		return game.AbilityContent{}, false
	}

	cond := c.Conditions[0]
	if cond.Kind != compiler.ConditionIf ||
		cond.Predicate != compiler.ConditionPredicatePriorInstructionAccepted ||
		cond.Negated {
		return game.AbilityContent{}, false
	}

	const resultKey = game.ResultKey("wheel-discarded-this-way")
	sequence := []game.Instruction{
		{
			Primitive:     game.Discard{EntireHand: true, Player: game.ControllerReference()},
			Optional:      true,
			PublishResult: resultKey,
		},
		{
			Primitive: game.Draw{
				Player: game.ControllerReference(),
				Amount: game.Dynamic(game.DynamicAmount{
					Kind:      game.DynamicAmountPreviousEffectResult,
					ResultKey: resultKey,
				}),
			},
			ResultGate: opt.Val(game.InstructionResultGate{
				Key:      resultKey,
				Accepted: game.TriTrue,
			}),
		},
	}
	return game.Mode{Text: ctx.text, Sequence: sequence}.Ability(), true
}
