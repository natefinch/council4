package c

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// ContaminantGrafter is the card definition for Contaminant Grafter.
//
// Type: Creature — Phyrexian Druid
// Cost: {4}{G}
//
// Oracle text:
//
//	Trample, toxic 1
//	Whenever one or more creatures you control deal combat damage to one or more players, proliferate.
//	Corrupted — At the beginning of your end step, if an opponent has three or more poison counters, draw a card, then you may put a land card from your hand onto the battlefield.
var ContaminantGrafter = newContaminantGrafter

func newContaminantGrafter() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Green),
		CardFace: game.CardFace{
			Name: "Contaminant Grafter",
			ManaCost: opt.Val(cost.Mana{
				cost.O(4),
				cost.G,
			}),
			Colors:    []color.Color{color.Green},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Phyrexian, types.Druid},
			Power:     opt.Val(game.PT{Value: 5}),
			Toughness: opt.Val(game.PT{Value: 5}),
			StaticAbilities: []game.StaticAbility{
				game.TrampleStaticBody,
				game.StaticAbility{
					KeywordAbilities: []game.KeywordAbility{
						game.ToxicKeyword{Amount: 1},
					},
				},
			},
			TriggeredAbilities: []game.TriggeredAbility{
				game.TriggeredAbility{
					Trigger: game.TriggerCondition{
						Type: game.TriggerWhenever,
						Pattern: game.TriggerPattern{
							Event:                 game.EventDamageDealt,
							Controller:            game.TriggerControllerYou,
							Subject:               game.TriggerSubjectDamageSource,
							OneOrMore:             true,
							RequireCombatDamage:   true,
							DamageRecipient:       game.DamageRecipientPlayer,
							DamageSourceSelection: game.Selection{RequiredTypes: []types.Card{types.Creature}},
						},
					},
					Content: game.Mode{
						Sequence: []game.Instruction{
							{
								Primitive: game.Proliferate{
									Amount: game.Fixed(1),
								},
							},
						},
					}.Ability(),
				},
				game.TriggeredAbility{
					Trigger: game.TriggerCondition{
						Type: game.TriggerAt,
						Pattern: game.TriggerPattern{
							Event:      game.EventBeginningOfStep,
							Controller: game.TriggerControllerYou,
							Step:       game.StepEnd,
						},
						InterveningIf: "if an opponent has three or more poison counters",
						InterveningCondition: opt.Val(game.Condition{
							AnyOpponentPoisonAtLeast: 3,
						}),
					},
					Content: game.Mode{
						Sequence: []game.Instruction{
							{
								Primitive: game.Draw{
									Amount: game.Fixed(1),
									Player: game.ControllerReference(),
								},
							},
							{
								Primitive: game.ChooseFromZone{
									Player:     game.ControllerReference(),
									SourceZone: zone.Hand,
									Filter:     game.Selection{RequiredTypes: []types.Card{types.Land}},
									Quantity:   game.Fixed(1),
									Destination: game.ChooseDestination{
										Zone: zone.Battlefield,
									},
									Prompt: "Choose a card to put onto the battlefield",
								},
								Optional: true,
							},
						},
					}.Ability(),
				},
			},
			OracleText: `
			Trample, toxic 1
			Whenever one or more creatures you control deal combat damage to one or more players, proliferate.
			Corrupted — At the beginning of your end step, if an opponent has three or more poison counters, draw a card, then you may put a land card from your hand onto the battlefield.
		`,
		},
	}
}
