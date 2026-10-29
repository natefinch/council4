package v

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// ViciousBattlerager is the card definition for Vicious Battlerager.
//
// Type: Creature — Dwarf Barbarian
// Cost: {3}{B}
//
// Oracle text:
//
//	When this creature enters, you take the initiative.
//	Spiked Retribution — Whenever this creature becomes blocked by a creature, that creature's controller loses 5 life.
var ViciousBattlerager = newViciousBattlerager

func newViciousBattlerager() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Black),
		CardFace: game.CardFace{
			Name: "Vicious Battlerager",
			ManaCost: opt.Val(cost.Mana{
				cost.O(3),
				cost.B,
			}),
			Colors:    []color.Color{color.Black},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Dwarf, types.Barbarian},
			Power:     opt.Val(game.PT{Value: 1}),
			Toughness: opt.Val(game.PT{Value: 5}),
			TriggeredAbilities: []game.TriggeredAbility{
				game.TriggeredAbility{
					Text: "When this creature enters, you take the initiative.",
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
								Sequence: []game.Instruction{
									game.Instruction{
										Primitive: game.TakeInitiative{
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
				game.TriggeredAbility{
					Text: "Spiked Retribution — Whenever this creature becomes blocked by a creature, that creature's controller loses 5 life.",
					Trigger: game.TriggerCondition{
						Type: game.TriggerWhenever,
						Pattern: game.TriggerPattern{
							Event:           game.EventAttackerBecameBlocked,
							BlockedScope:    game.BlockedTriggerEachBlocker,
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
										Primitive: game.LoseLife{
											Amount: game.Fixed(5),
											Player: game.ObjectControllerReference(game.EventRelatedPermanentReference()),
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
			When this creature enters, you take the initiative.
			Spiked Retribution — Whenever this creature becomes blocked by a creature, that creature's controller loses 5 life.
		`,
		},
	}
}
