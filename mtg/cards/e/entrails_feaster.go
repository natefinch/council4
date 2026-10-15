package e

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// EntrailsFeaster is the card definition for Entrails Feaster.
//
// Type: Creature — Zombie Cat
// Cost: {B}
//
// Oracle text:
//
//	At the beginning of your upkeep, you may exile a creature card from a graveyard. If you do, put a +1/+1 counter on this creature. If you don't, tap this creature.
var EntrailsFeaster = newEntrailsFeaster

func newEntrailsFeaster() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Black),
		CardFace: game.CardFace{
			Name: "Entrails Feaster",
			ManaCost: opt.Val(cost.Mana{
				cost.B,
			}),
			Colors:    []color.Color{color.Black},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Zombie, types.Cat},
			Power:     opt.Val(game.PT{Value: 1}),
			Toughness: opt.Val(game.PT{Value: 1}),
			TriggeredAbilities: []game.TriggeredAbility{
				game.TriggeredAbility{
					Text: "At the beginning of your upkeep, you may exile a creature card from a graveyard. If you do, put a +1/+1 counter on this creature. If you don't, tap this creature.",
					Trigger: game.TriggerCondition{
						Type: game.TriggerAt,
						Pattern: game.TriggerPattern{
							Event:           game.EventBeginningOfStep,
							Controller:      game.TriggerControllerYou,
							StackObjectKind: game.StackSpell,
							CounterKind:     counter.PlusOnePlusOne,
							Step:            game.StepUpkeep,
						},
					},
					Content: game.AbilityContent{
						Modes: []game.Mode{
							game.Mode{
								Sequence: []game.Instruction{
									game.Instruction{
										Primitive: game.ChooseFromZone{
											Player:     game.ControllerReference(),
											SourceZone: zone.Graveyard,
											AllOwners:  true,
											Filter:     game.Selection{RequiredTypes: []types.Card{types.Creature}},
											Quantity:   game.Fixed(1),
											Destination: game.ChooseDestination{
												Zone: zone.Exile,
											},
											Riders: game.ChooseRiders{
												PublishLinked: game.LinkedKey("exiled-with-source"),
											},
											Prompt: "Choose a card to exile",
										},
										Optional:      true,
										PublishResult: game.ResultKey("if-you-do"),
										LocalProducts: game.LocalProducts{
											Results: []game.ResultKey{game.ResultKey("if-you-do")},
										},
									},
									game.Instruction{
										Primitive: game.AddCounter{
											Amount:      game.Fixed(1),
											Object:      game.SourcePermanentReference(),
											CounterKind: counter.PlusOnePlusOne,
										},
										ResultGate: opt.Val(game.InstructionResultGate{
											Key:       game.ResultKey("if-you-do"),
											Succeeded: game.TriTrue,
										}),
									},
									game.Instruction{
										Primitive: game.Tap{
											Object: game.SourcePermanentReference(),
										},
										ResultGate: opt.Val(game.InstructionResultGate{
											Key:       game.ResultKey("if-you-do"),
											Succeeded: game.TriFalse,
										}),
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
			At the beginning of your upkeep, you may exile a creature card from a graveyard. If you do, put a +1/+1 counter on this creature. If you don't, tap this creature.
		`,
		},
	}
}
