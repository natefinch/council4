package s

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// SparringGolem is the card definition for Sparring Golem.
//
// Type: Artifact Creature — Golem
// Cost: {3}
//
// Oracle text:
//
//	Whenever this creature becomes blocked, it gets +1/+1 until end of turn for each creature blocking it.
var SparringGolem = newSparringGolem

func newSparringGolem() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name: "Sparring Golem",
			ManaCost: opt.Val(cost.Mana{
				cost.O(3),
			}),
			Types:     []types.Card{types.Artifact, types.Creature},
			Subtypes:  []types.Sub{types.Golem},
			Power:     opt.Val(game.PT{Value: 2}),
			Toughness: opt.Val(game.PT{Value: 2}),
			TriggeredAbilities: []game.TriggeredAbility{
				game.TriggeredAbility{
					Text: "Whenever this creature becomes blocked, it gets +1/+1 until end of turn for each creature blocking it.",
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
											Object: game.EventPermanentReference(),
											PowerDelta: game.Dynamic(game.DynamicAmount{
												Kind:       game.DynamicAmountBlockingCreatures,
												Multiplier: 1,
												Object:     game.EventPermanentReference(),
											}),
											ToughnessDelta: game.Dynamic(game.DynamicAmount{
												Kind:       game.DynamicAmountBlockingCreatures,
												Multiplier: 1,
												Object:     game.EventPermanentReference(),
											}),
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
			Whenever this creature becomes blocked, it gets +1/+1 until end of turn for each creature blocking it.
		`,
		},
	}
}
