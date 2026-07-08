package d

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// DistortedCuriosity is the card definition for Distorted Curiosity.
//
// Type: Sorcery
// Cost: {2}{U}
//
// Oracle text:
//
//	Corrupted — This spell costs {2} less to cast if an opponent has three or more poison counters.
//	Draw two cards.
var DistortedCuriosity = newDistortedCuriosity

func newDistortedCuriosity() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue),
		CardFace: game.CardFace{
			Name: "Distorted Curiosity",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
				cost.U,
			}),
			Colors: []color.Color{color.Blue},
			Types:  []types.Card{types.Sorcery},
			StaticAbilities: []game.StaticAbility{
				game.StaticAbility{
					RuleEffects: []game.RuleEffect{
						game.RuleEffect{
							Kind:           game.RuleEffectCostModifier,
							AffectedSource: true,
							CostModifier: game.CostModifier{
								Kind:             game.CostModifierSpell,
								GenericReduction: 2,
								ReductionCondition: opt.Val(game.Condition{
									AnyOpponentPoisonAtLeast: 3,
								}),
							},
						},
					},
				},
			},
			SpellAbility: opt.Val(game.Mode{
				Sequence: []game.Instruction{
					{
						Primitive: game.Draw{
							Amount: game.Fixed(2),
							Player: game.ControllerReference(),
						},
					},
				},
			}.Ability()),
			OracleText: `
			Corrupted — This spell costs {2} less to cast if an opponent has three or more poison counters.
			Draw two cards.
		`,
		},
	}
}
