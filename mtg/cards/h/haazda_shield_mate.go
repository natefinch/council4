package h

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// HaazdaShieldMate is the card definition for Haazda Shield Mate.
//
// Type: Creature — Human Soldier
// Cost: {2}{W}
//
// Oracle text:
//
//	At the beginning of your upkeep, sacrifice this creature unless you pay {W}{W}.
//	{W}: The next time a source of your choice would deal damage to you this turn, prevent that damage.
var HaazdaShieldMate = newHaazdaShieldMate

func newHaazdaShieldMate() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White),
		CardFace: game.CardFace{
			Name: "Haazda Shield Mate",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
				cost.W,
			}),
			Colors:    []color.Color{color.White},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Human, types.Soldier},
			Power:     opt.Val(game.PT{Value: 1}),
			Toughness: opt.Val(game.PT{Value: 1}),
			ActivatedAbilities: []game.ActivatedAbility{
				game.ActivatedAbility{
					Text:           "{W}: The next time a source of your choice would deal damage to you this turn, prevent that damage.",
					ManaCost:       opt.Val(cost.Mana{cost.W}),
					ZoneOfFunction: zone.Battlefield,
					Content: game.Mode{
						Sequence: []game.Instruction{
							{
								Primitive: game.PreventDamage{
									Player:  game.ControllerReference(),
									All:     true,
									OneShot: true,
								},
							},
						},
					}.Ability(),
				},
			},
			TriggeredAbilities: []game.TriggeredAbility{
				game.TriggeredAbility{
					Trigger: game.TriggerCondition{
						Type: game.TriggerAt,
						Pattern: game.TriggerPattern{
							Event:      game.EventBeginningOfStep,
							Controller: game.TriggerControllerYou,
							Step:       game.StepUpkeep,
						},
					},
					Content: game.Mode{
						Sequence: []game.Instruction{
							{
								Primitive: game.Pay{
									Payment: game.ResolutionPayment{
										Prompt: "Pay {W}{W}?",
										ManaCost: opt.Val(cost.Mana{
											cost.W,
											cost.W,
										}),
									},
								},
								PublishResult: game.ResultKey("sacrifice-unless-paid"),
							},
							{
								Primitive: game.Sacrifice{
									Object: game.SourcePermanentReference(),
								},
								ResultGate: opt.Val(game.InstructionResultGate{
									Key:       "sacrifice-unless-paid",
									Succeeded: game.TriFalse,
								}),
							},
						},
					}.Ability(),
				},
			},
			OracleText: `
			At the beginning of your upkeep, sacrifice this creature unless you pay {W}{W}.
			{W}: The next time a source of your choice would deal damage to you this turn, prevent that damage.
		`,
		},
	}
}
