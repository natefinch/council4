package c

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// ConstrictingSliver is the card definition for Constricting Sliver.
//
// Type: Creature — Sliver
// Cost: {5}{W}
//
// Oracle text:
//
//	Sliver creatures you control have "When this creature enters, you may exile target creature an opponent controls until this creature leaves the battlefield."
var ConstrictingSliver = newConstrictingSliver

func newConstrictingSliver() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White),
		CardFace: game.CardFace{
			Name: "Constricting Sliver",
			ManaCost: opt.Val(cost.Mana{
				cost.O(5),
				cost.W,
			}),
			Colors:    []color.Color{color.White},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Sliver},
			Power:     opt.Val(game.PT{Value: 3}),
			Toughness: opt.Val(game.PT{Value: 3}),
			StaticAbilities: []game.StaticAbility{
				game.StaticAbility{
					Text: "Sliver creatures you control have \"When this creature enters, you may exile target creature an opponent controls until this creature leaves the battlefield.\"",
					ContinuousEffects: []game.ContinuousEffect{
						game.ContinuousEffect{
							Duration: game.DurationPermanent,
							Group:    game.ObjectControlledGroup(game.SourcePermanentReference(), game.Selection{SubtypesAny: []types.Sub{types.Sub("Sliver")}}),
							Layer:    game.LayerAbility,
							AddAbilities: []game.Ability{
								&game.TriggeredAbility{
									Text: "When this creature enters, you may exile target creature an opponent controls until this creature leaves the battlefield.",
									Trigger: game.TriggerCondition{
										Type: game.TriggerWhen,
										Pattern: game.TriggerPattern{
											Event:           game.EventPermanentEnteredBattlefield,
											Source:          game.TriggerSourceSelf,
											StackObjectKind: game.StackSpell,
											CounterKind:     counter.PlusOnePlusOne,
										},
									},
									Optional: true,
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
														Primitive: game.MovePermanent{
															Object:        game.TargetPermanentReference(0),
															Destination:   zone.Exile,
															PublishLinked: game.LinkedKey("exile-until-leaves"),
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
			Sliver creatures you control have "When this creature enters, you may exile target creature an opponent controls until this creature leaves the battlefield."
		`,
		},
	}
}
