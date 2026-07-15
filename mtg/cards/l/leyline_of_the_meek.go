package l

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// LeylineOfTheMeek is the card definition for Leyline of the Meek.
//
// Type: Enchantment
// Cost: {2}{W}{W}
//
// Oracle text:
//
//	If this card is in your opening hand, you may begin the game with it on the battlefield.
//	Creature tokens get +1/+1.
var LeylineOfTheMeek = newLeylineOfTheMeek

func newLeylineOfTheMeek() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White),
		CardFace: game.CardFace{
			Name: "Leyline of the Meek",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
				cost.W,
				cost.W,
			}),
			Colors: []color.Color{color.White},
			Types:  []types.Card{types.Enchantment},
			StaticAbilities: []game.StaticAbility{
				game.StaticAbility{
					BeginsGameOnBattlefield: true,
				},
				game.StaticAbility{
					ContinuousEffects: []game.ContinuousEffect{
						game.ContinuousEffect{
							Layer:          game.LayerPowerToughnessModify,
							Group:          game.BattlefieldGroup(game.Selection{RequiredTypes: []types.Card{types.Creature}, TokenOnly: true}),
							PowerDelta:     1,
							ToughnessDelta: 1,
						},
					},
				},
			},
			OracleText: `
			If this card is in your opening hand, you may begin the game with it on the battlefield.
			Creature tokens get +1/+1.
		`,
		},
	}
}
