package h

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// HeartlessSummoning is the card definition for Heartless Summoning.
//
// Type: Enchantment
// Cost: {1}{B}
//
// Oracle text:
//
//	Creature spells you cast cost {2} less to cast.
//	Creatures you control get -1/-1.
var HeartlessSummoning = newHeartlessSummoning

func newHeartlessSummoning() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Black),
		CardFace: game.CardFace{
			Name: "Heartless Summoning",
			ManaCost: opt.Val(cost.Mana{
				cost.O(1),
				cost.B,
			}),
			Colors: []color.Color{color.Black},
			Types:  []types.Card{types.Enchantment},
			StaticAbilities: []game.StaticAbility{
				game.StaticAbility{
					Text: "Creature spells you cast cost {2} less to cast.",
					RuleEffects: []game.RuleEffect{
						game.RuleEffect{
							Kind:           game.RuleEffectCostModifier,
							Duration:       game.DurationPermanent,
							AffectedPlayer: game.PlayerYou,
							CostModifier: game.CostModifier{
								Kind:             game.CostModifierSpell,
								GenericReduction: 2,
								CardSelection:    game.Selection{RequiredTypes: []types.Card{types.Creature}},
							},
						},
					},
				},
				game.StaticAbility{
					Text: "Creatures you control get -1/-1.",
					ContinuousEffects: []game.ContinuousEffect{
						game.ContinuousEffect{
							Duration:       game.DurationPermanent,
							Group:          game.ObjectControlledGroup(game.SourcePermanentReference(), game.Selection{RequiredTypes: []types.Card{types.Creature}}),
							Layer:          game.LayerPowerToughnessModify,
							PowerDelta:     -1,
							ToughnessDelta: -1,
						},
					},
				},
			},
			OracleText: `
			Creature spells you cast cost {2} less to cast.
			Creatures you control get -1/-1.
		`,
		},
	}
}
