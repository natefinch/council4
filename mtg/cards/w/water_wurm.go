package w

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// WaterWurm is the card definition for Water Wurm.
//
// Type: Creature — Wurm
// Cost: {U}
//
// Oracle text:
//
//	This creature gets +0/+1 as long as an opponent controls an Island.
var WaterWurm = newWaterWurm

func newWaterWurm() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue),
		CardFace: game.CardFace{
			Name: "Water Wurm",
			ManaCost: opt.Val(cost.Mana{
				cost.U,
			}),
			Colors:    []color.Color{color.Blue},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Wurm},
			Power:     opt.Val(game.PT{Value: 1}),
			Toughness: opt.Val(game.PT{Value: 1}),
			StaticAbilities: []game.StaticAbility{
				game.StaticAbility{
					Text: "This creature gets +0/+1 as long as an opponent controls an Island.",
					Condition: opt.Val(game.Condition{
						Text: "as long as an opponent controls an Island",
						AnyOpponentControls: opt.Val(game.SelectionCount{
							Selection: game.Selection{SubtypesAny: []types.Sub{types.Island}},
						}),
						SourceCounterKind: counter.PlusOnePlusOne,
					}),
					ContinuousEffects: []game.ContinuousEffect{
						game.ContinuousEffect{
							Duration:       game.DurationPermanent,
							AffectedSource: true,
							Layer:          game.LayerPowerToughnessModify,
							ToughnessDelta: 1,
						},
					},
				},
			},
			OracleText: `
			This creature gets +0/+1 as long as an opponent controls an Island.
		`,
		},
	}
}
