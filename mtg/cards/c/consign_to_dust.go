package c

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// ConsignToDust is the card definition for Consign to Dust.
//
// Type: Instant
// Cost: {2}{G}
//
// Oracle text:
//
//	Strive — This spell costs {2}{G} more to cast for each target beyond the first.
//	Destroy any number of target artifacts and/or enchantments.
var ConsignToDust = newConsignToDust

func newConsignToDust() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Green),
		CardFace: game.CardFace{
			Name: "Consign to Dust",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
				cost.G,
			}),
			Colors: []color.Color{color.Green},
			Types:  []types.Card{types.Instant},
			StaticAbilities: []game.StaticAbility{
				game.StaticAbility{
					Text: "Strive — This spell costs {2}{G} more to cast for each target beyond the first.",
					RuleEffects: []game.RuleEffect{
						game.RuleEffect{
							Kind:           game.RuleEffectCostModifier,
							AffectedSource: true,
							Duration:       game.DurationPermanent,
							CostModifier: game.CostModifier{
								Kind: game.CostModifierSpell,
								PerTargetBeyondFirstIncrease: cost.Mana{
									cost.O(2),
									cost.G,
								},
							},
						},
					},
				},
			},
			SpellAbility: opt.Val(game.Mode{
				Targets: []game.TargetSpec{
					game.TargetSpec{
						MinTargets: 0,
						MaxTargets: 99,
						Constraint: "any number of target artifacts and/or enchantments",
						Allow:      game.TargetAllowPermanent,
						Selection:  opt.Val(game.Selection{RequiredTypesAny: []types.Card{types.Artifact, types.Enchantment}}),
					},
				},
				Sequence: []game.Instruction{
					{
						Primitive: game.Destroy{
							Object: game.AllTargetPermanentsReference(0),
						},
					},
				},
			}.Ability()),
			OracleText: `
			Strive — This spell costs {2}{G} more to cast for each target beyond the first.
			Destroy any number of target artifacts and/or enchantments.
		`,
		},
	}
}
