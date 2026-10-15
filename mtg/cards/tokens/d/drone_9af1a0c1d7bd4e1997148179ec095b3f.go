package d

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// Drone is the card definition for Drone.
//
// Type: Token Artifact Creature — Drone
//
// Oracle text:
//   Deathtouch
//   When this creature leaves the battlefield, each opponent loses 2 life and you gain 2 life.

// DroneToken9af1a0c1d7bd4e1997148179ec095b3f is the card definition for Drone.
var DroneToken9af1a0c1d7bd4e1997148179ec095b3f = newDroneToken9af1a0c1d7bd4e1997148179ec095b3f()

func newDroneToken9af1a0c1d7bd4e1997148179ec095b3f() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name:      "Drone",
			Types:     []types.Card{types.Artifact, types.Creature},
			Subtypes:  []types.Sub{types.Drone},
			Power:     opt.Val(game.PT{Value: 2}),
			Toughness: opt.Val(game.PT{Value: 2}),
			StaticAbilities: []game.StaticAbility{
				game.DeathtouchStaticBody,
			},
			TriggeredAbilities: []game.TriggeredAbility{
				game.TriggeredAbility{
					Text: "When this creature leaves the battlefield, each opponent loses 2 life and you gain 2 life.",
					Trigger: game.TriggerCondition{
						Type: game.TriggerWhen,
						Pattern: game.TriggerPattern{
							Event:           game.EventZoneChanged,
							Source:          game.TriggerSourceSelf,
							MatchFromZone:   true,
							FromZone:        zone.Battlefield,
							StackObjectKind: game.StackSpell,
							CounterKind:     counter.PlusOnePlusOne,
						},
					},
					Content: game.AbilityContent{
						Modes: []game.Mode{
							game.Mode{
								Sequence: []game.Instruction{
									game.Instruction{
										Primitive: game.LoseLife{
											Amount:      game.Fixed(2),
											PlayerGroup: game.OpponentsReference(),
										},
										PublishResult: game.ResultKey("life-change"),
									},
									game.Instruction{
										Primitive: game.GainLife{
											Amount: game.Fixed(2),
											Player: game.ControllerReference(),
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
			Deathtouch
			When this creature leaves the battlefield, each opponent loses 2 life and you gain 2 life.
		`,
		},
	}
}
