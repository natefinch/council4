package s

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// SunspearShikari is the card definition for Sunspear Shikari.
//
// Type: Creature — Cat Soldier
// Cost: {1}{W}
//
// Oracle text:
//
//	As long as this creature is equipped, it has first strike and lifelink. (It deals combat damage before creatures without first strike. Damage dealt by it also causes you to gain that much life.)
var SunspearShikari = newSunspearShikari

func newSunspearShikari() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White),
		CardFace: game.CardFace{
			Name: "Sunspear Shikari",
			ManaCost: opt.Val(cost.Mana{
				cost.O(1),
				cost.W,
			}),
			Colors:    []color.Color{color.White},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Cat, types.Soldier},
			Power:     opt.Val(game.PT{Value: 2}),
			Toughness: opt.Val(game.PT{Value: 2}),
			StaticAbilities: []game.StaticAbility{
				game.StaticAbility{
					Condition: opt.Val(game.Condition{
						Object:        opt.Val(game.SourcePermanentReference()),
						ObjectMatches: opt.Val(game.Selection{RequiredTypes: []types.Card{types.Creature}, MatchEquipped: true}),
					}),
					ContinuousEffects: []game.ContinuousEffect{
						game.ContinuousEffect{
							Layer:          game.LayerAbility,
							AffectedSource: true,
							AddKeywords: []game.Keyword{
								game.FirstStrike,
								game.Lifelink,
							},
						},
					},
				},
			},
			OracleText: `
			As long as this creature is equipped, it has first strike and lifelink. (It deals combat damage before creatures without first strike. Damage dealt by it also causes you to gain that much life.)
		`,
		},
	}
}
