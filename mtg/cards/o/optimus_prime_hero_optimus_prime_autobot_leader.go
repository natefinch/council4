package o

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// OptimusPrimeHero is the card definition for Optimus Prime, Hero // Optimus Prime, Autobot Leader.
//
// Type: Legendary Artifact Creature — Robot // Legendary Artifact — Vehicle
// Face: Optimus Prime, Autobot Leader — Legendary Artifact — Vehicle
//
// Oracle text:
//
//	More Than Meets the Eye {2}{U}{R}{W} (You may cast this card converted for {2}{U}{R}{W}.)
//	At the beginning of each end step, bolster 1. (Choose a creature with the least toughness among creatures you control and put a +1/+1 counter on it.)
//	When Optimus Prime dies, return it to the battlefield converted under its owner's control.
var OptimusPrimeHero = newOptimusPrimeHero

func newOptimusPrimeHero() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White, color.Blue, color.Red),
		CardFace: game.CardFace{
			Name: "Optimus Prime, Hero",
			ManaCost: opt.Val(cost.Mana{
				cost.O(3),
				cost.U,
				cost.R,
				cost.W,
			}),
			Colors:     []color.Color{color.Red, color.Blue, color.White},
			Supertypes: []types.Super{types.Legendary},
			Types:      []types.Card{types.Artifact, types.Creature},
			Subtypes:   []types.Sub{types.Robot},
			Power:      opt.Val(game.PT{Value: 4}),
			Toughness:  opt.Val(game.PT{Value: 8}),
			TriggeredAbilities: []game.TriggeredAbility{
				game.TriggeredAbility{
					Trigger: game.TriggerCondition{
						Type: game.TriggerAt,
						Pattern: game.TriggerPattern{
							Event: game.EventBeginningOfStep,
							Step:  game.StepEnd,
						},
					},
					Content: game.Mode{
						Sequence: []game.Instruction{
							{
								Primitive: game.Bolster{
									Amount: game.Fixed(1),
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
								Primitive: game.PutOnBattlefield{
									Source:           game.CardBattlefieldSource(game.CardReference{Kind: game.CardReferenceEvent}),
									EntryTransformed: true,
								},
							},
						},
					}.Ability(),
				},
			},
			AlternativeCosts: []cost.Alternative{
				cost.Alternative{
					Label:    "More Than Meets the Eye",
					ManaCost: opt.Val(cost.Mana{cost.O(2), cost.U, cost.R, cost.W}),
					Mechanic: cost.AlternativeMechanicMoreThanMeetsTheEye,
				},
			},
			OracleText: `
			More Than Meets the Eye {2}{U}{R}{W} (You may cast this card converted for {2}{U}{R}{W}.)
			At the beginning of each end step, bolster 1. (Choose a creature with the least toughness among creatures you control and put a +1/+1 counter on it.)
			When Optimus Prime dies, return it to the battlefield converted under its owner's control.
		`,
		},
		Layout: game.LayoutTransform,
		Back: opt.Val(game.CardFace{
			Name:       "Optimus Prime, Autobot Leader",
			Colors:     []color.Color{color.Red, color.Blue, color.White},
			Supertypes: []types.Super{types.Legendary},
			Types:      []types.Card{types.Artifact},
			Subtypes:   []types.Sub{types.Vehicle},
			Power:      opt.Val(game.PT{Value: 6}),
			Toughness:  opt.Val(game.PT{Value: 8}),
			StaticAbilities: []game.StaticAbility{
				game.LivingMetalStaticBody,
				game.TrampleStaticBody,
			},
			TriggeredAbilities: []game.TriggeredAbility{
				game.TriggeredAbility{
					Trigger: game.TriggerCondition{
						Type: game.TriggerWhenever,
						Pattern: game.TriggerPattern{
							Event:      game.EventAttackerDeclared,
							Controller: game.TriggerControllerYou,
							OneOrMore:  true,
						},
					},
					Content: game.Mode{
						Sequence: []game.Instruction{
							{
								Primitive: game.Bolster{
									Amount:        game.Fixed(2),
									PublishLinked: game.LinkedKey("bolster-chosen"),
								},
							},
							{
								Primitive: game.ApplyContinuous{
									Object: opt.Val(game.LinkedObjectReference("bolster-chosen")),
									ContinuousEffects: []game.ContinuousEffect{
										game.ContinuousEffect{
											Layer: game.LayerAbility,
											AddKeywords: []game.Keyword{
												game.Trample,
											},
										},
									},
									Duration: game.DurationUntilEndOfTurn,
								},
							},
							{
								Primitive: game.CreateDelayedTrigger{
									Trigger: game.DelayedTriggerDef{
										EventPattern: opt.Val(game.TriggerPattern{
											Event:                game.EventDamageDealt,
											RequireCombatDamage:  true,
											DamageRecipient:      game.DamageRecipientPlayer,
											DamageSourceCaptured: true,
										}),
										Window:             game.DelayedWindowThisTurn,
										DamageSourceObject: opt.Val(game.LinkedObjectReference("bolster-chosen")),
										Content: game.Mode{
											Sequence: []game.Instruction{
												{
													Primitive: game.Transform{
														Object: game.SourcePermanentReference(),
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
			Living metal (During your turn, this Vehicle is also a creature.)
			Trample
			Whenever you attack, bolster 2. The chosen creature gains trample until end of turn. When that creature deals combat damage to a player this turn, convert Optimus Prime.
		`,
		}),
	}
}
