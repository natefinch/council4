package cardgen

import (
	"github.com/natefinch/council4/cardgen/oracle/compiler"
	"github.com/natefinch/council4/mtg/game"
)

// lowerBottomHandThenDrawSequence lowers the parser-recognized spell body "put
// any number of cards from your hand on the {bottom|top} of your library, then
// draw that many cards[ plus N]" into its fixed instruction template. The
// compiler marks the body with a text-blind exact-sequence kind and carries the
// library end and draw offset as typed parameters; this lowering reads only
// those typed values, so it never inspects Oracle words.
func lowerBottomHandThenDrawSequence(ability compiler.CompiledAbility) game.AbilityContent {
	sequence := []game.Instruction{
		{
			Primitive: game.PutHandOnLibraryThenDraw{
				Player:     game.ControllerReference(),
				Bottom:     ability.ExactSequenceBottom,
				DrawOffset: int(ability.ExactSequenceDrawOffset),
			},
		},
	}
	return game.Mode{Text: ability.Text, Sequence: sequence}.Ability()
}
