package u

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// UginSInsight is the card definition for Ugin's Insight.
//
// Type: Sorcery
// Cost: {3}{U}{U}
//
// Oracle text:
//
//	Scry X, where X is the greatest mana value among permanents you control, then draw three cards.
var UginSInsight = newUginSInsight()

func newUginSInsight() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue),
		CardFace: game.CardFace{
			Name: "Ugin's Insight",
			ManaCost: opt.Val(cost.Mana{
				cost.O(3),
				cost.U,
				cost.U,
			}),
			Colors: []color.Color{color.Blue},
			Types:  []types.Card{types.Sorcery},
			SpellAbility: opt.Val(game.Mode{
				Sequence: []game.Instruction{
					{
						Primitive: game.Scry{
							Amount: game.Dynamic(game.DynamicAmount{
								Kind:       game.DynamicAmountGreatestManaValueInGroup,
								Multiplier: 1,
								Group:      game.BattlefieldGroup(game.Selection{Controller: game.ControllerYou}),
							}),
							Player: game.ControllerReference(),
						},
					},
					{
						Primitive: game.Draw{
							Amount: game.Fixed(3),
							Player: game.ControllerReference(),
						},
					},
				},
			}.Ability()),
			OracleText: `
			Scry X, where X is the greatest mana value among permanents you control, then draw three cards.
		`,
		},
	}
}
