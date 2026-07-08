package m

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// MesmerizingDose is the card definition for Mesmerizing Dose.
//
// Type: Enchantment — Aura
// Cost: {1}{U}{U}
//
// Oracle text:
//
//	Enchant creature
//	When this Aura enters, tap enchanted creature, then proliferate. (Choose any number of permanents and/or players, then give each another counter of each kind already there.)
//	Enchanted creature doesn't untap during its controller's untap step.
var MesmerizingDose = newMesmerizingDose

func newMesmerizingDose() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue),
		CardFace: game.CardFace{
			Name: "Mesmerizing Dose",
			ManaCost: opt.Val(cost.Mana{
				cost.O(1),
				cost.U,
				cost.U,
			}),
			Colors:   []color.Color{color.Blue},
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
							Kind:             game.RuleEffectDoesntUntap,
							AffectedAttached: true,
						},
					},
				},
			},
			TriggeredAbilities: []game.TriggeredAbility{
				game.TriggeredAbility{
					Trigger: game.TriggerCondition{
						Type: game.TriggerWhen,
						Pattern: game.TriggerPattern{
							Event:  game.EventPermanentEnteredBattlefield,
							Source: game.TriggerSourceSelf,
						},
					},
					Content: game.Mode{
						Sequence: []game.Instruction{
							{
								Primitive: game.Tap{
									Object: game.SourceAttachedPermanentReference(),
								},
							},
							{
								Primitive: game.Proliferate{
									Amount: game.Fixed(1),
								},
							},
						},
					}.Ability(),
				},
			},
			OracleText: `
			Enchant creature
			When this Aura enters, tap enchanted creature, then proliferate. (Choose any number of permanents and/or players, then give each another counter of each kind already there.)
			Enchanted creature doesn't untap during its controller's untap step.
		`,
		},
	}
}
