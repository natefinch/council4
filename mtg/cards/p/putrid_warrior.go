package p

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// PutridWarrior is the card definition for Putrid Warrior.
//
// Type: Creature — Zombie Soldier Warrior
// Cost: {W}{B}
//
// Oracle text:
//
//	Whenever this creature deals damage, choose one —
//	• Each player loses 1 life.
//	• Each player gains 1 life.
var PutridWarrior = newPutridWarrior

func newPutridWarrior() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White, color.Black),
		CardFace: game.CardFace{
			Name: "Putrid Warrior",
			ManaCost: opt.Val(cost.Mana{
				cost.W,
				cost.B,
			}),
			Colors:    []color.Color{color.Black, color.White},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Zombie, types.Soldier, types.Warrior},
			Power:     opt.Val(game.PT{Value: 2}),
			Toughness: opt.Val(game.PT{Value: 2}),
			TriggeredAbilities: []game.TriggeredAbility{
				game.TriggeredAbility{
					Trigger: game.TriggerCondition{
						Type: game.TriggerWhenever,
						Pattern: game.TriggerPattern{
							Event:   game.EventDamageDealt,
							Source:  game.TriggerSourceSelf,
							Subject: game.TriggerSubjectDamageSource,
						},
					},
					Content: game.AbilityContent{
						Modes: []game.Mode{
							game.Mode{
								Text: "Each player loses 1 life.",
								Sequence: []game.Instruction{
									{
										Primitive: game.LoseLife{
											Amount:      game.Fixed(1),
											PlayerGroup: game.AllPlayersReference(),
										},
									},
								},
							},
							game.Mode{
								Text: "Each player gains 1 life.",
								Sequence: []game.Instruction{
									{
										Primitive: game.GainLife{
											Amount:      game.Fixed(1),
											PlayerGroup: game.AllPlayersReference(),
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
			Whenever this creature deals damage, choose one —
			• Each player loses 1 life.
			• Each player gains 1 life.
		`,
		},
	}
}
