package s

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// SnubhornSentry is the card definition for Snubhorn Sentry.
//
// Type: Creature — Dinosaur
// Cost: {W}
//
// Oracle text:
//
//	Ascend (If you control ten or more permanents, you get the city's blessing for the rest of the game.)
//	This creature gets +3/+0 as long as you have the city's blessing.
var SnubhornSentry = newSnubhornSentry

func newSnubhornSentry() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White),
		CardFace: game.CardFace{
			Name: "Snubhorn Sentry",
			ManaCost: opt.Val(cost.Mana{
				cost.W,
			}),
			Colors:    []color.Color{color.White},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Dinosaur},
			Power:     opt.Val(game.PT{Value: 0}),
			Toughness: opt.Val(game.PT{Value: 3}),
			StaticAbilities: []game.StaticAbility{
				game.AscendStaticBody,
				game.StaticAbility{
					Condition: opt.Val(game.Condition{
						ControllerHasCityBlessing: true,
					}),
					ContinuousEffects: []game.ContinuousEffect{
						game.ContinuousEffect{
							Layer:          game.LayerPowerToughnessModify,
							AffectedSource: true,
							PowerDelta:     3,
						},
					},
				},
			},
			OracleText: `
			Ascend (If you control ten or more permanents, you get the city's blessing for the rest of the game.)
			This creature gets +3/+0 as long as you have the city's blessing.
		`,
		},
	}
}
