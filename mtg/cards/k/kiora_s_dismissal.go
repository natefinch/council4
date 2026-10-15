package k

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// KioraSDismissal is the card definition for Kiora's Dismissal.
//
// Type: Instant
// Cost: {U}
//
// Oracle text:
//
//	Strive — This spell costs {U} more to cast for each target beyond the first.
//	Return any number of target enchantments to their owners' hands.
var KioraSDismissal = newKioraSDismissal

func newKioraSDismissal() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue),
		CardFace: game.CardFace{
			Name: "Kiora's Dismissal",
			ManaCost: opt.Val(cost.Mana{
				cost.U,
			}),
			Colors: []color.Color{color.Blue},
			Types:  []types.Card{types.Instant},
			StaticAbilities: []game.StaticAbility{
				game.StaticAbility{
					Text: "Strive — This spell costs {U} more to cast for each target beyond the first.",
					RuleEffects: []game.RuleEffect{
						game.RuleEffect{
							Kind:           game.RuleEffectCostModifier,
							AffectedSource: true,
							Duration:       game.DurationPermanent,
							CostModifier: game.CostModifier{
								Kind: game.CostModifierSpell,
								PerTargetBeyondFirstIncrease: cost.Mana{
									cost.U,
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
						Constraint: "any number of target enchantments",
						Allow:      game.TargetAllowPermanent,
						Selection:  opt.Val(game.Selection{RequiredTypesAny: []types.Card{types.Enchantment}}),
					},
				},
				Sequence: []game.Instruction{
					{
						Primitive: game.MovePermanent{
							Object:      game.AllTargetPermanentsReference(0),
							Destination: zone.Hand,
						},
					},
				},
			}.Ability()),
			OracleText: `
			Strive — This spell costs {U} more to cast for each target beyond the first.
			Return any number of target enchantments to their owners' hands.
		`,
		},
	}
}
