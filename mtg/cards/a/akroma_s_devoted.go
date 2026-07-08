package a

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// AkromaSDevoted is the card definition for Akroma's Devoted.
//
// Type: Creature — Human Cleric
// Cost: {3}{W}
//
// Oracle text:
//
//	Cleric creatures have vigilance.
var AkromaSDevoted = newAkromaSDevoted

func newAkromaSDevoted() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White),
		CardFace: game.CardFace{
			Name: "Akroma's Devoted",
			ManaCost: opt.Val(cost.Mana{
				cost.O(3),
				cost.W,
			}),
			Colors:    []color.Color{color.White},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Human, types.Cleric},
			Power:     opt.Val(game.PT{Value: 2}),
			Toughness: opt.Val(game.PT{Value: 4}),
			StaticAbilities: []game.StaticAbility{
				game.StaticAbility{
					ContinuousEffects: []game.ContinuousEffect{
						game.ContinuousEffect{
							Layer: game.LayerAbility,
							Group: game.BattlefieldGroup(game.Selection{SubtypesAny: []types.Sub{types.Sub("Cleric")}}),
							AddKeywords: []game.Keyword{
								game.Vigilance,
							},
						},
					},
				},
			},
			OracleText: `
			Cleric creatures have vigilance.
		`,
		},
	}
}
