package t

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// TauntingSliver is the card definition for Taunting Sliver.
//
// Type: Creature — Sliver
// Cost: {3}{U}
//
// Oracle text:
//
//	Sliver creatures you control have "When this creature enters, goad target creature an opponent controls." (Until your next turn, that creature attacks each combat if able and attacks a player other than you if able.)
var TauntingSliver = newTauntingSliver

func newTauntingSliver() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue),
		CardFace: game.CardFace{
			Name: "Taunting Sliver",
			ManaCost: opt.Val(cost.Mana{
				cost.O(3),
				cost.U,
			}),
			Colors:    []color.Color{color.Blue},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Sliver},
			Power:     opt.Val(game.PT{Value: 3}),
			Toughness: opt.Val(game.PT{Value: 3}),
			StaticAbilities: []game.StaticAbility{
				game.StaticAbility{
					Text: "Sliver creatures you control have \"When this creature enters, goad target creature an opponent controls.\" (Until your next turn, that creature attacks each combat if able and attacks a player other than you if able.)",
					ContinuousEffects: []game.ContinuousEffect{
						game.ContinuousEffect{
							Duration: game.DurationPermanent,
							Group:    game.ObjectControlledGroup(game.SourcePermanentReference(), game.Selection{SubtypesAny: []types.Sub{types.Sub("Sliver")}}),
							Layer:    game.LayerAbility,
							AddAbilities: []game.Ability{
								&game.TriggeredAbility{
									Text: "When this creature enters, goad target creature an opponent controls.",
									Trigger: game.TriggerCondition{
										Type: game.TriggerWhen,
										Pattern: game.TriggerPattern{
											Event:           game.EventPermanentEnteredBattlefield,
											Source:          game.TriggerSourceSelf,
											StackObjectKind: game.StackSpell,
											CounterKind:     counter.PlusOnePlusOne,
										},
									},
									Content: game.AbilityContent{
										Modes: []game.Mode{
											game.Mode{
												Targets: []game.TargetSpec{
													game.TargetSpec{
														MinTargets: 1,
														MaxTargets: 1,
														Constraint: "target creature an opponent controls",
														Allow:      game.TargetAllowPermanent,
														Selection:  opt.Val(game.Selection{RequiredTypesAny: []types.Card{types.Creature}, Controller: game.ControllerOpponent}),
													},
												},
												Sequence: []game.Instruction{
													game.Instruction{
														Primitive: game.Goad{
															Object: game.TargetPermanentReference(0),
														},
													},
												},
											},
										},
										MinModes: 1,
										MaxModes: 1,
									},
								},
							},
						},
					},
				},
			},
			OracleText: `
			Sliver creatures you control have "When this creature enters, goad target creature an opponent controls." (Until your next turn, that creature attacks each combat if able and attacks a player other than you if able.)
		`,
		},
	}
}
