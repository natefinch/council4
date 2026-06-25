package g

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// GiantSGrasp is the card definition for Giant's Grasp.
//
// Type: Enchantment — Aura
// Cost: {2}{U}{U}
//
// Oracle text:
//
//	Enchant Giant you control
//	When this Aura enters, gain control of target nonland permanent for as long as this Aura remains on the battlefield.
var GiantSGrasp = newGiantSGrasp()

func newGiantSGrasp() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue),
		CardFace: game.CardFace{
			Name: "Giant's Grasp",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
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
					Constraint: "giant you control",
					Allow:      game.TargetAllowPermanent,
					Selection: opt.Val(game.Selection{
						SubtypesAny: []types.Sub{types.Sub("Giant")},
						Controller:  game.ControllerYou,
					}),
				}),
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
						Targets: []game.TargetSpec{
							game.TargetSpec{
								MinTargets: 1,
								MaxTargets: 1,
								Constraint: "target nonland permanent",
								Allow:      game.TargetAllowPermanent,
								Selection: opt.Val(game.Selection{
									ExcludedTypes: []types.Card{types.Land},
								}),
							},
						},
						Sequence: []game.Instruction{
							{
								Primitive: game.ApplyContinuous{
									Object: opt.Val(game.TargetPermanentReference(0)),
									ContinuousEffects: []game.ContinuousEffect{
										game.ContinuousEffect{
											Layer:         game.LayerControl,
											NewController: opt.Val(game.Player1),
										},
									},
									Duration: game.DurationForAsLongAsSourceOnBattlefield,
								},
							},
						},
					}.Ability(),
				},
			},
			OracleText: `
			Enchant Giant you control
			When this Aura enters, gain control of target nonland permanent for as long as this Aura remains on the battlefield.
		`,
		},
	}
}
