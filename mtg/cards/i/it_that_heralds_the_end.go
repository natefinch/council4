package i

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/compare"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// ItThatHeraldsTheEnd is the card definition for It That Heralds the End.
//
// Type: Creature — Eldrazi Drone
// Cost: {1}{C}
//
// Oracle text:
//
//	Colorless spells you cast with mana value 7 or greater cost {1} less to cast.
//	Other colorless creatures you control get +1/+1.
var ItThatHeraldsTheEnd = newItThatHeraldsTheEnd()

func newItThatHeraldsTheEnd() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name: "It That Heralds the End",
			ManaCost: opt.Val(cost.Mana{
				cost.O(1),
				cost.C,
			}),
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Eldrazi, types.Drone},
			Power:     opt.Val(game.PT{Value: 2}),
			Toughness: opt.Val(game.PT{Value: 2}),
			StaticAbilities: []game.StaticAbility{
				game.StaticAbility{
					RuleEffects: []game.RuleEffect{
						game.RuleEffect{
							Kind:           game.RuleEffectCostModifier,
							AffectedPlayer: game.PlayerYou,
							CostModifier: game.CostModifier{
								Kind:             game.CostModifierSpell,
								CardSelection:    game.Selection{Colorless: true, ManaValue: opt.Val(compare.Int{Op: compare.GreaterOrEqual, Value: 7})},
								GenericReduction: 1,
							},
						},
					},
				},
				game.StaticAbility{
					ContinuousEffects: []game.ContinuousEffect{
						game.ContinuousEffect{
							Layer:          game.LayerPowerToughnessModify,
							Group:          game.ObjectControlledGroupExcluding(game.SourcePermanentReference(), game.Selection{RequiredTypes: []types.Card{types.Creature}, Colorless: true}, game.SourcePermanentReference()),
							PowerDelta:     1,
							ToughnessDelta: 1,
						},
					},
				},
			},
			OracleText: `
			Colorless spells you cast with mana value 7 or greater cost {1} less to cast.
			Other colorless creatures you control get +1/+1.
		`,
		},
	}
}
