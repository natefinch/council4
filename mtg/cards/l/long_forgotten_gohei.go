package l

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// LongForgottenGohei is the card definition for Long-Forgotten Gohei.
//
// Type: Artifact
// Cost: {3}
//
// Oracle text:
//
//	Arcane spells you cast cost {1} less to cast.
//	Spirit creatures you control get +1/+1.
var LongForgottenGohei = newLongForgottenGohei

func newLongForgottenGohei() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name: "Long-Forgotten Gohei",
			ManaCost: opt.Val(cost.Mana{
				cost.O(3),
			}),
			Types: []types.Card{types.Artifact},
			StaticAbilities: []game.StaticAbility{
				game.StaticAbility{
					Text: "Arcane spells you cast cost {1} less to cast.",
					RuleEffects: []game.RuleEffect{
						game.RuleEffect{
							Kind:           game.RuleEffectCostModifier,
							Duration:       game.DurationPermanent,
							AffectedPlayer: game.PlayerYou,
							CostModifier: game.CostModifier{
								Kind:             game.CostModifierSpell,
								GenericReduction: 1,
								CardSelection:    game.Selection{SubtypesAny: []types.Sub{types.Arcane}},
							},
						},
					},
				},
				game.StaticAbility{
					Text: "Spirit creatures you control get +1/+1.",
					ContinuousEffects: []game.ContinuousEffect{
						game.ContinuousEffect{
							Duration:       game.DurationPermanent,
							Group:          game.ObjectControlledGroup(game.SourcePermanentReference(), game.Selection{SubtypesAny: []types.Sub{types.Sub("Spirit")}}),
							Layer:          game.LayerPowerToughnessModify,
							PowerDelta:     1,
							ToughnessDelta: 1,
						},
					},
				},
			},
			OracleText: `
			Arcane spells you cast cost {1} less to cast.
			Spirit creatures you control get +1/+1.
		`,
		},
	}
}
