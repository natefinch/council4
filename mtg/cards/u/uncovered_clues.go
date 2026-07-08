package u

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// UncoveredClues is the card definition for Uncovered Clues.
//
// Type: Sorcery
// Cost: {2}{U}
//
// Oracle text:
//
//	Look at the top four cards of your library. You may reveal up to two instant and/or sorcery cards from among them and put the revealed cards into your hand. Put the rest on the bottom of your library in any order.
var UncoveredClues = newUncoveredClues

func newUncoveredClues() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue),
		CardFace: game.CardFace{
			Name: "Uncovered Clues",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
				cost.U,
			}),
			Colors: []color.Color{color.Blue},
			Types:  []types.Card{types.Sorcery},
			SpellAbility: opt.Val(game.Mode{
				Sequence: []game.Instruction{
					{
						Primitive: game.Dig{
							Player:    game.ControllerReference(),
							Look:      game.Fixed(4),
							Take:      game.Fixed(2),
							Remainder: game.DigRemainderLibraryBottom,
							Filter:    opt.Val(game.Selection{RequiredTypesAny: []types.Card{types.Instant, types.Sorcery}}),
							TakeUpTo:  true,
							Reveal:    true,
						},
					},
				},
			}.Ability()),
			OracleText: `
			Look at the top four cards of your library. You may reveal up to two instant and/or sorcery cards from among them and put the revealed cards into your hand. Put the rest on the bottom of your library in any order.
		`,
		},
	}
}
