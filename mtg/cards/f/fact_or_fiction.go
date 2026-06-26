package f

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// FactOrFiction is the card definition for Fact or Fiction.
//
// Type: Instant
// Cost: {3}{U}
//
// Oracle text:
//
//	Reveal the top five cards of your library. An opponent separates those cards into two piles. Put one pile into your hand and the other into your graveyard.
var FactOrFiction = newFactOrFiction()

func newFactOrFiction() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue),
		CardFace: game.CardFace{
			Name: "Fact or Fiction",
			ManaCost: opt.Val(cost.Mana{
				cost.O(3),
				cost.U,
			}),
			Colors: []color.Color{color.Blue},
			Types:  []types.Card{types.Instant},
			SpellAbility: opt.Val(game.Mode{
				Sequence: []game.Instruction{
					{
						Primitive: game.PileSplit{
							Player:            game.ControllerReference(),
							Amount:            game.Fixed(5),
							SeparatorOpponent: true,
							Kept:              zone.Hand,
							Other:             zone.Graveyard,
						},
					},
				},
			}.Ability()),
			OracleText: `
			Reveal the top five cards of your library. An opponent separates those cards into two piles. Put one pile into your hand and the other into your graveyard.
		`,
		},
	}
}
