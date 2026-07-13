package s

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// SkymarcherAspirant is the card definition for Skymarcher Aspirant.
//
// Type: Creature — Vampire Soldier
// Cost: {W}
//
// Oracle text:
//
//	Ascend (If you control ten or more permanents, you get the city's blessing for the rest of the game.)
//	This creature has flying as long as you have the city's blessing.
var SkymarcherAspirant = newSkymarcherAspirant

func newSkymarcherAspirant() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White),
		CardFace: game.CardFace{
			Name: "Skymarcher Aspirant",
			ManaCost: opt.Val(cost.Mana{
				cost.W,
			}),
			Colors:    []color.Color{color.White},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Vampire, types.Soldier},
			Power:     opt.Val(game.PT{Value: 2}),
			Toughness: opt.Val(game.PT{Value: 1}),
			StaticAbilities: []game.StaticAbility{
				game.AscendStaticBody,
				game.StaticAbility{
					Condition: opt.Val(game.Condition{
						ControllerHasCityBlessing: true,
					}),
					ContinuousEffects: []game.ContinuousEffect{
						game.ContinuousEffect{
							Layer:          game.LayerAbility,
							AffectedSource: true,
							AddKeywords: []game.Keyword{
								game.Flying,
							},
						},
					},
				},
			},
			OracleText: `
			Ascend (If you control ten or more permanents, you get the city's blessing for the rest of the game.)
			This creature has flying as long as you have the city's blessing.
		`,
		},
	}
}
