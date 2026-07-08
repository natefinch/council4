package j

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// JunkGolem is the card definition for Junk Golem.
//
// Type: Artifact Creature — Golem
// Cost: {4}
//
// Oracle text:
//
//	This creature enters with three +1/+1 counters on it.
//	At the beginning of your upkeep, sacrifice this creature unless you remove a +1/+1 counter from it.
//	{1}, Discard a card: Put a +1/+1 counter on this creature.
var JunkGolem = newJunkGolem

func newJunkGolem() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name: "Junk Golem",
			ManaCost: opt.Val(cost.Mana{
				cost.O(4),
			}),
			Types:     []types.Card{types.Artifact, types.Creature},
			Subtypes:  []types.Sub{types.Golem},
			Power:     opt.Val(game.PT{Value: 0}),
			Toughness: opt.Val(game.PT{Value: 0}),
			ActivatedAbilities: []game.ActivatedAbility{
				game.ActivatedAbility{
					Text:     "{1}, Discard a card: Put a +1/+1 counter on this creature.",
					ManaCost: opt.Val(cost.Mana{cost.O(1)}),
					AdditionalCosts: []cost.Additional{
						{
							Kind:   cost.AdditionalDiscard,
							Text:   "Discard a card",
							Amount: 1,
							Source: zone.Hand,
						},
					},
					ZoneOfFunction: zone.Battlefield,
					Content: game.Mode{
						Sequence: []game.Instruction{
							{
								Primitive: game.AddCounter{
									Amount:      game.Fixed(1),
									Object:      game.SourcePermanentReference(),
									CounterKind: counter.PlusOnePlusOne,
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
										Prompt: "Remove a +1/+1 counter from it?",
										AdditionalCosts: []cost.Additional{
											{
												Kind:        cost.AdditionalRemoveCounter,
												Text:        "remove a +1/+1 counter from it",
												Amount:      1,
												CounterKind: counter.PlusOnePlusOne,
											},
										},
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
			ReplacementAbilities: []game.ReplacementAbility{
				game.EntersWithCountersReplacement("This creature enters with three +1/+1 counters on it.", game.CounterPlacement{Kind: counter.PlusOnePlusOne, Amount: 3}),
			},
			OracleText: `
			This creature enters with three +1/+1 counters on it.
			At the beginning of your upkeep, sacrifice this creature unless you remove a +1/+1 counter from it.
			{1}, Discard a card: Put a +1/+1 counter on this creature.
		`,
		},
	}
}
