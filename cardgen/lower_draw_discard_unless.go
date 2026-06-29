package cardgen

import (
	"github.com/natefinch/council4/cardgen/oracle/compiler"
	"github.com/natefinch/council4/mtg/game"
)

// lowerDrawThenDiscardUnlessSequence lowers the parser-recognized spell body
// "Draw N cards. Then discard M cards unless you discard a <type[ or type...]>
// card." (Thirst for Knowledge family) into a fixed Draw followed by a
// DiscardUnlessType: the controller draws ExactSequenceDrawCount cards, then
// discards ExactSequenceDiscardCount cards unless they instead discard a single
// card of one of the recorded exempt types. The compiler carries the counts and
// exempt types as typed values, so this lowering reads no Oracle words.
func lowerDrawThenDiscardUnlessSequence(ability compiler.CompiledAbility) game.AbilityContent {
	sequence := []game.Instruction{
		{
			Primitive: game.Draw{
				Player: game.ControllerReference(),
				Amount: game.Fixed(int(ability.ExactSequenceDrawCount)),
			},
		},
		{
			Primitive: game.DiscardUnlessType{
				Player:      game.ControllerReference(),
				Amount:      int(ability.ExactSequenceDiscardCount),
				ExemptTypes: ability.ExactSequenceLookAtTopTypes,
			},
		},
	}
	return game.Mode{Text: ability.Text, Sequence: sequence}.Ability()
}
