package h

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// Hibernation is the card definition for Hibernation.
//
// Type: Instant
// Cost: {2}{U}
//
// Oracle text:
//
//	Return all green permanents to their owners' hands.
var Hibernation = newHibernation

func newHibernation() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue),
		CardFace: game.CardFace{
			Name: "Hibernation",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
				cost.U,
			}),
			Colors: []color.Color{color.Blue},
			Types:  []types.Card{types.Instant},
			SpellAbility: opt.Val(game.Mode{
				Sequence: []game.Instruction{
					{
						Primitive: game.MovePermanent{
							Group:       game.BattlefieldGroup(game.Selection{ColorsAny: []color.Color{color.Green}}),
							Destination: zone.Hand,
						},
					},
				},
			}.Ability()),
			OracleText: `
			Return all green permanents to their owners' hands.
		`,
		},
	}
}
