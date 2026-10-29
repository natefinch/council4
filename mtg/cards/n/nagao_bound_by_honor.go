package n

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// NagaoBoundByHonor is the card definition for Nagao, Bound by Honor.
//
// Type: Legendary Creature — Human Samurai
// Cost: {3}{W}
//
// Oracle text:
//
//	Bushido 1 (Whenever this creature blocks or becomes blocked, it gets +1/+1 until end of turn.)
//	Whenever Nagao attacks, Samurai creatures you control get +1/+1 until end of turn.
var NagaoBoundByHonor = newNagaoBoundByHonor

func newNagaoBoundByHonor() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White),
		CardFace: game.CardFace{
			Name: "Nagao, Bound by Honor",
			ManaCost: opt.Val(cost.Mana{
				cost.O(3),
				cost.W,
			}),
			Colors:     []color.Color{color.White},
			Supertypes: []types.Super{types.Legendary},
			Types:      []types.Card{types.Creature},
			Subtypes:   []types.Sub{types.Human, types.Samurai},
			Power:      opt.Val(game.PT{Value: 3}),
			Toughness:  opt.Val(game.PT{Value: 3}),
			TriggeredAbilities: []game.TriggeredAbility{
				game.TriggeredAbility{
					Text: "Whenever this creature blocks or becomes blocked, it gets +1/+1 until end of turn.",
					Trigger: game.TriggerCondition{
						Type: game.TriggerWhenever,
						Pattern: game.TriggerPattern{
							Event:           game.EventBlockerDeclared,
							BlockerScope:    game.BlockerTriggerOnce,
							Source:          game.TriggerSourceSelf,
							StackObjectKind: game.StackSpell,
							CounterKind:     counter.PlusOnePlusOne,
							UnionEvent:      game.EventAttackerBecameBlocked,
							BlockedScope:    game.BlockedTriggerOnce,
						},
					},
					Content: game.AbilityContent{
						Modes: []game.Mode{
							game.Mode{
								Sequence: []game.Instruction{
									game.Instruction{
										Primitive: game.ModifyPT{
											Object:         game.EventPermanentReference(),
											PowerDelta:     game.Fixed(1),
											ToughnessDelta: game.Fixed(1),
											Duration:       game.DurationUntilEndOfTurn,
										},
									},
								},
							},
						},
						MinModes: 1,
						MaxModes: 1,
					},
				},
				game.TriggeredAbility{
					Text: "Whenever Nagao attacks, Samurai creatures you control get +1/+1 until end of turn.",
					Trigger: game.TriggerCondition{
						Type: game.TriggerWhenever,
						Pattern: game.TriggerPattern{
							Event:           game.EventAttackerDeclared,
							Source:          game.TriggerSourceSelf,
							StackObjectKind: game.StackSpell,
							CounterKind:     counter.PlusOnePlusOne,
						},
					},
					Content: game.AbilityContent{
						Modes: []game.Mode{
							game.Mode{
								Sequence: []game.Instruction{
									game.Instruction{
										Primitive: game.ApplyContinuous{
											ContinuousEffects: []game.ContinuousEffect{
												game.ContinuousEffect{
													Duration:       game.DurationPermanent,
													Group:          game.BattlefieldGroup(game.Selection{SubtypesAny: []types.Sub{types.Sub("Samurai")}, Controller: game.ControllerYou}),
													Layer:          game.LayerPowerToughnessModify,
													PowerDelta:     1,
													ToughnessDelta: 1,
												},
											},
											Duration: game.DurationUntilEndOfTurn,
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
			OracleText: `
			Bushido 1 (Whenever this creature blocks or becomes blocked, it gets +1/+1 until end of turn.)
			Whenever Nagao attacks, Samurai creatures you control get +1/+1 until end of turn.
		`,
		},
	}
}
