package a

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// AnabaSpiritCrafter is the card definition for Anaba Spirit Crafter.
//
// Type: Creature — Minotaur Shaman
// Cost: {2}{R}{R}
//
// Oracle text:
//
//	Minotaur creatures get +1/+0.
var AnabaSpiritCrafter = newAnabaSpiritCrafter

func newAnabaSpiritCrafter() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Red),
		CardFace: game.CardFace{
			Name: "Anaba Spirit Crafter",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
				cost.R,
				cost.R,
			}),
			Colors:    []color.Color{color.Red},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Minotaur, types.Shaman},
			Power:     opt.Val(game.PT{Value: 1}),
			Toughness: opt.Val(game.PT{Value: 3}),
			StaticAbilities: []game.StaticAbility{
				game.StaticAbility{
					ContinuousEffects: []game.ContinuousEffect{
						game.ContinuousEffect{
							Layer:      game.LayerPowerToughnessModify,
							Group:      game.BattlefieldGroup(game.Selection{SubtypesAny: []types.Sub{types.Sub("Minotaur")}}),
							PowerDelta: 1,
						},
					},
				},
			},
			OracleText: `
			Minotaur creatures get +1/+0.
		`,
		},
	}
}
