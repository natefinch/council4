package cardgen

import (
	"github.com/natefinch/council4/cardgen/oracle/compiler"
	"github.com/natefinch/council4/mtg/game"
)

// lowerDiscardHandThenDrawSequence lowers the parser-recognized spell body
// "Discard {your hand | all the cards in your hand}, then draw that many cards."
// (Decaying Time Loop) into a whole-hand Discard followed by a Draw that reads
// the published discard count. The discard publishes the number of cards the
// controller discarded under a result key, and the draw's dynamic amount reads
// that key, so the controller draws exactly as many cards as they discarded.
// The compiler marks the body with a text-blind exact-sequence kind, so this
// lowering reads no Oracle words.
func lowerDiscardHandThenDrawSequence(ability compiler.CompiledAbility) game.AbilityContent {
	const resultKey = game.ResultKey("discarded-this-way")
	sequence := []game.Instruction{
		{
			Primitive:     game.Discard{EntireHand: true, Player: game.ControllerReference()},
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
		},
	}
	return game.Mode{Text: ability.Text, Sequence: sequence}.Ability()
}
