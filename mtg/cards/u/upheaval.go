package u

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// Upheaval is the card definition for Upheaval.
//
// Type: Sorcery
// Cost: {4}{U}{U}
//
// Oracle text:
//
//	Return all permanents to their owners' hands.
var Upheaval = newUpheaval

func newUpheaval() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue),
		CardFace: game.CardFace{
			Name: "Upheaval",
			ManaCost: opt.Val(cost.Mana{
				cost.O(4),
				cost.U,
				cost.U,
			}),
			Colors: []color.Color{color.Blue},
			Types:  []types.Card{types.Sorcery},
			SpellAbility: opt.Val(game.Mode{
				Sequence: []game.Instruction{
					{
						Primitive: game.MovePermanent{
							Group:       game.BattlefieldGroup(game.Selection{}),
							Destination: zone.Hand,
						},
					},
				},
			}.Ability()),
			OracleText: `
			Return all permanents to their owners' hands.
		`,
		},
	}
}
