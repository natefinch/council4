package i

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// IchorclawMyr is the card definition for Ichorclaw Myr.
//
// Type: Artifact Creature — Phyrexian Myr
// Cost: {2}
//
// Oracle text:
//
//	Infect (This creature deals damage to creatures in the form of -1/-1 counters and to players in the form of poison counters.)
//	Whenever this creature becomes blocked, it gets +2/+2 until end of turn.
var IchorclawMyr = newIchorclawMyr

func newIchorclawMyr() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name: "Ichorclaw Myr",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
			}),
			Types:     []types.Card{types.Artifact, types.Creature},
			Subtypes:  []types.Sub{types.Phyrexian, types.Myr},
			Power:     opt.Val(game.PT{Value: 1}),
			Toughness: opt.Val(game.PT{Value: 1}),
			StaticAbilities: []game.StaticAbility{
				game.InfectStaticBody,
			},
			TriggeredAbilities: []game.TriggeredAbility{
				game.TriggeredAbility{
					Text: "Whenever this creature becomes blocked, it gets +2/+2 until end of turn.",
					Trigger: game.TriggerCondition{
						Type: game.TriggerWhenever,
						Pattern: game.TriggerPattern{
							Event:           game.EventAttackerBecameBlocked,
							BlockedScope:    game.BlockedTriggerOnce,
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
										Primitive: game.ModifyPT{
											Object:         game.EventPermanentReference(),
											PowerDelta:     game.Fixed(2),
											ToughnessDelta: game.Fixed(2),
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
			},
			OracleText: `
			Infect (This creature deals damage to creatures in the form of -1/-1 counters and to players in the form of poison counters.)
			Whenever this creature becomes blocked, it gets +2/+2 until end of turn.
		`,
		},
	}
}
