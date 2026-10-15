package t

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// TimeWipe is the card definition for Time Wipe.
//
// Type: Sorcery
// Cost: {2}{W}{W}{U}
//
// Oracle text:
//
//	Return a creature you control to its owner's hand, then destroy all creatures.
var TimeWipe = newTimeWipe

func newTimeWipe() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White, color.Blue),
		CardFace: game.CardFace{
			Name: "Time Wipe",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
				cost.W,
				cost.W,
				cost.U,
			}),
			Colors: []color.Color{color.Blue, color.White},
			Types:  []types.Card{types.Sorcery},
			SpellAbility: opt.Val(game.Mode{
				Sequence: []game.Instruction{
					{
						Primitive: game.MovePermanent{
							Group:            game.BattlefieldGroup(game.Selection{RequiredTypes: []types.Card{types.Creature}, Controller: game.ControllerYou}),
							Destination:      zone.Hand,
							ControlledChoice: true,
							Amount:           game.Fixed(1),
						},
					},
					{
						Primitive: game.Destroy{
							Group: game.BattlefieldGroup(game.Selection{RequiredTypes: []types.Card{types.Creature}}),
						},
					},
				},
			}.Ability()),
			OracleText: `
			Return a creature you control to its owner's hand, then destroy all creatures.
		`,
		},
	}
}
