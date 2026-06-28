package c

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// ChokingRestraints is the card definition for Choking Restraints.
//
// Type: Enchantment — Aura
// Cost: {2}{W}
//
// Oracle text:
//
//	Enchant creature
//	Enchanted creature can't attack or block.
//	{3}{W}{W}, Sacrifice this Aura: Exile enchanted creature.
var ChokingRestraints = newChokingRestraints()

func newChokingRestraints() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White),
		CardFace: game.CardFace{
			Name: "Choking Restraints",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
				cost.W,
			}),
			Colors:   []color.Color{color.White},
			Types:    []types.Card{types.Enchantment},
			Subtypes: []types.Sub{types.Aura},
			StaticAbilities: []game.StaticAbility{
				game.EnchantStaticAbility(&game.TargetSpec{
					MinTargets: 1,
					MaxTargets: 1,
					Constraint: "creature",
					Allow:      game.TargetAllowPermanent,
					Selection:  opt.Val(game.Selection{RequiredTypesAny: []types.Card{types.Creature}}),
				}),
				game.StaticAbility{
					RuleEffects: []game.RuleEffect{
						game.RuleEffect{
							Kind:             game.RuleEffectCantAttack,
							AffectedAttached: true,
						},
						game.RuleEffect{
							Kind:             game.RuleEffectCantBlock,
							AffectedAttached: true,
						},
					},
				},
			},
			ActivatedAbilities: []game.ActivatedAbility{
				game.ActivatedAbility{
					Text:     "{3}{W}{W}, Sacrifice this Aura: Exile enchanted creature.",
					ManaCost: opt.Val(cost.Mana{cost.O(3), cost.W, cost.W}),
					AdditionalCosts: []cost.Additional{
						{
							Kind:   cost.AdditionalSacrificeSource,
							Text:   "Sacrifice this Aura",
							Amount: 1,
						},
					},
					ZoneOfFunction: zone.Battlefield,
					Content: game.Mode{
						Sequence: []game.Instruction{
							{
								Primitive: game.Exile{
									Object: game.SourceAttachedPermanentReference(),
								},
							},
						},
					}.Ability(),
				},
			},
			OracleText: `
			Enchant creature
			Enchanted creature can't attack or block.
			{3}{W}{W}, Sacrifice this Aura: Exile enchanted creature.
		`,
		},
	}
}
