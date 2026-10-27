// Legacy regression fixture preserved when the Oracle compiler was retired.
package fixtures

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// RTLibrarySpell is the card definition for RT Library Spell.
//
// Type: Sorcery
//
// Oracle text:
//
//	Look at the top card of your library.
var RTLibrarySpell = newRTLibrarySpell

func newRTLibrarySpell() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name:  "RT Library Spell",
			Types: []types.Card{types.Sorcery},
			SpellAbility: opt.Val(game.Mode{
				Sequence: []game.Instruction{
					{
						Primitive: game.LookAtLibraryTop{
							Player:        game.ControllerReference(),
							PublishLinked: game.LinkedKey("sequence-effect-0-product"),
						},
						LocalProducts: game.LocalProducts{
							Links: []game.LinkedKey{game.LinkedKey("sequence-effect-0-product")},
						},
					},
				},
			}.Ability()),
			OracleText: `
			Look at the top card of your library.
		`,
		},
	}
}
