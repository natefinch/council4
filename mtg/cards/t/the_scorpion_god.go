package t

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// TheScorpionGod is the card definition for The Scorpion God.
//
// Type: Legendary Creature — God
// Cost: {3}{B}{R}
//
// Oracle text:
//
//	Whenever a creature with a -1/-1 counter on it dies, draw a card.
//	{1}{B}{R}: Put a -1/-1 counter on another target creature.
//	When The Scorpion God dies, return it to its owner's hand at the beginning of the next end step.
var TheScorpionGod = newTheScorpionGod

func newTheScorpionGod() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Black, color.Red),
		CardFace: game.CardFace{
			Name: "The Scorpion God",
			ManaCost: opt.Val(cost.Mana{
				cost.O(3),
				cost.B,
				cost.R,
			}),
			Colors:     []color.Color{color.Black, color.Red},
			Supertypes: []types.Super{types.Legendary},
			Types:      []types.Card{types.Creature},
			Subtypes:   []types.Sub{types.God},
			Power:      opt.Val(game.PT{Value: 6}),
			Toughness:  opt.Val(game.PT{Value: 5}),
			ActivatedAbilities: []game.ActivatedAbility{
				game.ActivatedAbility{
					Text:           "{1}{B}{R}: Put a -1/-1 counter on another target creature.",
					ManaCost:       opt.Val(cost.Mana{cost.O(1), cost.B, cost.R}),
					ZoneOfFunction: zone.Battlefield,
					Content: game.Mode{
						Targets: []game.TargetSpec{
							game.TargetSpec{
								MinTargets: 1,
								MaxTargets: 1,
								Constraint: "another target creature",
								Allow:      game.TargetAllowPermanent,
								Selection:  opt.Val(game.Selection{RequiredTypesAny: []types.Card{types.Creature}, ExcludeSource: true}),
							},
						},
						Sequence: []game.Instruction{
							{
								Primitive: game.AddCounter{
									Amount:      game.Fixed(1),
									Object:      game.TargetPermanentReference(0),
									CounterKind: counter.MinusOneMinusOne,
								},
							},
						},
					}.Ability(),
				},
			},
			TriggeredAbilities: []game.TriggeredAbility{
				game.TriggeredAbility{
					Trigger: game.TriggerCondition{
						Type: game.TriggerWhenever,
						Pattern: game.TriggerPattern{
							Event:            game.EventPermanentDied,
							SubjectSelection: game.Selection{RequiredTypes: []types.Card{types.Creature}, MatchCounter: true, RequiredCounter: counter.MinusOneMinusOne},
						},
					},
					Content: game.Mode{
						Sequence: []game.Instruction{
							{
								Primitive: game.Draw{
									Amount: game.Fixed(1),
									Player: game.ControllerReference(),
								},
							},
						},
					}.Ability(),
				},
				game.TriggeredAbility{
					Trigger: game.TriggerCondition{
						Type: game.TriggerWhen,
						Pattern: game.TriggerPattern{
							Event:            game.EventPermanentDied,
							Source:           game.TriggerSourceSelf,
							SubjectSelection: game.Selection{RequiredTypes: []types.Card{types.Creature}},
						},
					},
					Content: game.Mode{
						Sequence: []game.Instruction{
							{
								Primitive: game.CreateDelayedTrigger{
									Trigger: game.DelayedTriggerDef{
										Timing: game.DelayedAtBeginningOfNextEndStep,
										Content: game.Mode{
											Sequence: []game.Instruction{
												{
													Primitive: game.MoveCard{
														Card:        game.CardReference{Kind: game.CardReferenceSource},
														FromZone:    zone.Graveyard,
														Destination: zone.Hand,
													},
												},
											},
										}.Ability(),
									},
								},
							},
						},
					}.Ability(),
				},
			},
			OracleText: `
			Whenever a creature with a -1/-1 counter on it dies, draw a card.
			{1}{B}{R}: Put a -1/-1 counter on another target creature.
			When The Scorpion God dies, return it to its owner's hand at the beginning of the next end step.
		`,
		},
	}
}
