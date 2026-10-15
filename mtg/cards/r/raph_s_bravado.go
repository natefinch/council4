package r

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// RaphSBravado is the card definition for Raph's Bravado.
//
// Type: Enchantment
// Cost: {1}{R}
//
// Oracle text:
//
//	During your turn, attacking creatures get +1/+0.
var RaphSBravado = newRaphSBravado

func newRaphSBravado() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Red),
		CardFace: game.CardFace{
			Name: "Raph's Bravado",
			ManaCost: opt.Val(cost.Mana{
				cost.O(1),
				cost.R,
			}),
			Colors: []color.Color{color.Red},
			Types:  []types.Card{types.Enchantment},
			StaticAbilities: []game.StaticAbility{
				game.StaticAbility{
					Text: "During your turn, attacking creatures get +1/+0.",
					Condition: opt.Val(game.Condition{
						Text:                 "During your turn",
						SourceCounterKind:    counter.PlusOnePlusOne,
						SourceControllerTurn: true,
					}),
					ContinuousEffects: []game.ContinuousEffect{
						game.ContinuousEffect{
							Duration:   game.DurationPermanent,
							Group:      game.BattlefieldGroup(game.Selection{RequiredTypes: []types.Card{types.Creature}, CombatState: game.CombatStateAttacking}),
							Layer:      game.LayerPowerToughnessModify,
							PowerDelta: 1,
						},
					},
				},
			},
			OracleText: `
			During your turn, attacking creatures get +1/+0.
		`,
		},
	}
}
