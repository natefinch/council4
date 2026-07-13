package s

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// SpireWinder is the card definition for Spire Winder.
//
// Type: Creature — Snake
// Cost: {3}{U}
//
// Oracle text:
//
//	Flying
//	Ascend (If you control ten or more permanents, you get the city's blessing for the rest of the game.)
//	This creature gets +1/+1 as long as you have the city's blessing.
var SpireWinder = newSpireWinder

func newSpireWinder() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue),
		CardFace: game.CardFace{
			Name: "Spire Winder",
			ManaCost: opt.Val(cost.Mana{
				cost.O(3),
				cost.U,
			}),
			Colors:    []color.Color{color.Blue},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Snake},
			Power:     opt.Val(game.PT{Value: 2}),
			Toughness: opt.Val(game.PT{Value: 3}),
			StaticAbilities: []game.StaticAbility{
				game.FlyingStaticBody,
				game.AscendStaticBody,
				game.StaticAbility{
					Condition: opt.Val(game.Condition{
						ControllerHasCityBlessing: true,
					}),
					ContinuousEffects: []game.ContinuousEffect{
						game.ContinuousEffect{
							Layer:          game.LayerPowerToughnessModify,
							AffectedSource: true,
							PowerDelta:     1,
							ToughnessDelta: 1,
						},
					},
				},
			},
			OracleText: `
			Flying
			Ascend (If you control ten or more permanents, you get the city's blessing for the rest of the game.)
			This creature gets +1/+1 as long as you have the city's blessing.
		`,
		},
	}
}
