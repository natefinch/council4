package r

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// RushOfKnowledge is the card definition for Rush of Knowledge.
//
// Type: Sorcery
// Cost: {4}{U}
//
// Oracle text:
//
//	Draw cards equal to the greatest mana value among permanents you control.
var RushOfKnowledge = newRushOfKnowledge

func newRushOfKnowledge() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue),
		CardFace: game.CardFace{
			Name: "Rush of Knowledge",
			ManaCost: opt.Val(cost.Mana{
				cost.O(4),
				cost.U,
			}),
			Colors: []color.Color{color.Blue},
			Types:  []types.Card{types.Sorcery},
			SpellAbility: opt.Val(game.Mode{
				Sequence: []game.Instruction{
					{
						Primitive: game.Draw{
							Amount: game.Dynamic(game.DynamicAmount{
								Kind:       game.DynamicAmountGreatestManaValueInGroup,
								Multiplier: 1,
								Group:      game.BattlefieldGroup(game.Selection{Controller: game.ControllerYou}),
							}),
							Player: game.ControllerReference(),
						},
					},
				},
			}.Ability()),
			OracleText: `
			Draw cards equal to the greatest mana value among permanents you control.
		`,
		},
	}
}
