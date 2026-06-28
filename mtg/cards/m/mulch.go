package m

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// Mulch is the card definition for Mulch.
//
// Type: Sorcery
// Cost: {1}{G}
//
// Oracle text:
//
//	Reveal the top four cards of your library. Put all land cards revealed this way into your hand and the rest into your graveyard.
var Mulch = newMulch()

func newMulch() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Green),
		CardFace: game.CardFace{
			Name: "Mulch",
			ManaCost: opt.Val(cost.Mana{
				cost.O(1),
				cost.G,
			}),
			Colors: []color.Color{color.Green},
			Types:  []types.Card{types.Sorcery},
			SpellAbility: opt.Val(game.Mode{
				Sequence: []game.Instruction{
					{
						Primitive: game.RevealTopPartition{
							Player:    game.ControllerReference(),
							Amount:    game.Fixed(4),
							Selection: game.Selection{RequiredTypes: []types.Card{types.Land}, Controller: game.ControllerYou},
						},
					},
				},
			}.Ability()),
			OracleText: `
			Reveal the top four cards of your library. Put all land cards revealed this way into your hand and the rest into your graveyard.
		`,
		},
	}
}
