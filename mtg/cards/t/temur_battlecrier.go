package t

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/compare"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// TemurBattlecrier is the card definition for Temur Battlecrier.
//
// Type: Creature — Orc Ranger
// Cost: {G}{U}{R}
//
// Oracle text:
//
//	During your turn, spells you cast cost {1} less to cast for each creature you control with power 4 or greater.
var TemurBattlecrier = newTemurBattlecrier

func newTemurBattlecrier() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue, color.Red, color.Green),
		CardFace: game.CardFace{
			Name: "Temur Battlecrier",
			ManaCost: opt.Val(cost.Mana{
				cost.G,
				cost.U,
				cost.R,
			}),
			Colors:    []color.Color{color.Green, color.Red, color.Blue},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Orc, types.Ranger},
			Power:     opt.Val(game.PT{Value: 4}),
			Toughness: opt.Val(game.PT{Value: 3}),
			StaticAbilities: []game.StaticAbility{
				game.StaticAbility{
					RuleEffects: []game.RuleEffect{
						game.RuleEffect{
							Kind:           game.RuleEffectCostModifier,
							AffectedPlayer: game.PlayerYou,
							CostModifier: game.CostModifier{
								Kind:               game.CostModifierSpell,
								PerObjectReduction: 1,
								CountSelection:     &game.Selection{RequiredTypes: []types.Card{types.Creature}, Controller: game.ControllerYou, Power: opt.Val(compare.Int{Op: compare.GreaterOrEqual, Value: 4})},
							},
							RestrictedDuringControllerTurn: true,
						},
					},
				},
			},
			OracleText: `
			During your turn, spells you cast cost {1} less to cast for each creature you control with power 4 or greater.
		`,
		},
	}
}
