package a

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// Apocalypse is the card definition for Apocalypse.
//
// Type: Sorcery
// Cost: {2}{R}{R}{R}
//
// Oracle text:
//
//	Exile all permanents. You discard your hand.
var Apocalypse = newApocalypse

func newApocalypse() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Red),
		CardFace: game.CardFace{
			Name: "Apocalypse",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
				cost.R,
				cost.R,
				cost.R,
			}),
			Colors: []color.Color{color.Red},
			Types:  []types.Card{types.Sorcery},
			SpellAbility: opt.Val(game.Mode{
				Sequence: []game.Instruction{
					{
						Primitive: game.MovePermanent{
							Group:       game.BattlefieldGroup(game.Selection{}),
							Destination: zone.Exile,
						},
					},
					{
						Primitive: game.Discard{
							Player:     game.ControllerReference(),
							EntireHand: true,
						},
					},
				},
			}.Ability()),
			OracleText: `
			Exile all permanents. You discard your hand.
		`,
		},
	}
}
