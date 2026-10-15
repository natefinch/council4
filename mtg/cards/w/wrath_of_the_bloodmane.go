package w

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// WrathOfTheBloodmane is the card definition for Wrath of the Bloodmane.
//
// Type: Instant
// Cost: {2}{R}
//
// Oracle text:
//
//	This spell costs {1} less to cast if you control a legendary creature.
//	Wrath of the Bloodmane deals 4 damage to target creature or planeswalker.
var WrathOfTheBloodmane = newWrathOfTheBloodmane

func newWrathOfTheBloodmane() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Red),
		CardFace: game.CardFace{
			Name: "Wrath of the Bloodmane",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
				cost.R,
			}),
			Colors: []color.Color{color.Red},
			Types:  []types.Card{types.Instant},
			StaticAbilities: []game.StaticAbility{
				game.StaticAbility{
					Text: "This spell costs {1} less to cast if you control a legendary creature.",
					RuleEffects: []game.RuleEffect{
						game.RuleEffect{
							Kind:           game.RuleEffectCostModifier,
							AffectedSource: true,
							Duration:       game.DurationPermanent,
							CostModifier: game.CostModifier{
								Kind:             game.CostModifierSpell,
								GenericReduction: 1,
								ReductionCondition: opt.Val(game.Condition{
									Text: "if you control a legendary creature",
									ControlsMatching: opt.Val(game.SelectionCount{
										Selection: game.Selection{RequiredTypes: []types.Card{types.Creature}, Supertypes: []types.Super{types.Legendary}},
									}),
									SourceCounterKind: counter.PlusOnePlusOne,
								}),
							},
						},
					},
				},
			},
			SpellAbility: opt.Val(game.Mode{
				Targets: []game.TargetSpec{
					game.TargetSpec{
						MinTargets: 1,
						MaxTargets: 1,
						Constraint: "target creature or planeswalker",
						Allow:      game.TargetAllowPermanent,
						Selection:  opt.Val(game.Selection{RequiredTypesAny: []types.Card{types.Creature, types.Planeswalker}}),
					},
				},
				Sequence: []game.Instruction{
					{
						Primitive: game.Damage{
							Amount:    game.Fixed(4),
							Recipient: game.AnyTargetDamageRecipient(0),
						},
					},
				},
			}.Ability()),
			OracleText: `
			This spell costs {1} less to cast if you control a legendary creature.
			Wrath of the Bloodmane deals 4 damage to target creature or planeswalker.
		`,
		},
	}
}
