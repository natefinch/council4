package e

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/compare"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// ErebosBleakHearted is the card definition for Erebos, Bleak-Hearted.
//
// Type: Legendary Enchantment Creature — God
// Cost: {3}{B}
//
// Oracle text:
//
//	Indestructible
//	As long as your devotion to black is less than five, Erebos isn't a creature.
//	Whenever another creature you control dies, you may pay 2 life. If you do, draw a card.
//	{1}{B}, Sacrifice another creature: Target creature gets -2/-1 until end of turn.
var ErebosBleakHearted = newErebosBleakHearted

func newErebosBleakHearted() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Black),
		CardFace: game.CardFace{
			Name: "Erebos, Bleak-Hearted",
			ManaCost: opt.Val(cost.Mana{
				cost.O(3),
				cost.B,
			}),
			Colors:     []color.Color{color.Black},
			Supertypes: []types.Super{types.Legendary},
			Types:      []types.Card{types.Enchantment, types.Creature},
			Subtypes:   []types.Sub{types.God},
			Power:      opt.Val(game.PT{Value: 5}),
			Toughness:  opt.Val(game.PT{Value: 6}),
			StaticAbilities: []game.StaticAbility{
				game.IndestructibleStaticBody,
				game.StaticAbility{
					Condition: opt.Val(game.Condition{
						Aggregates: []game.AggregateComparison{{Aggregate: game.AggregateControllerDevotion, Op: compare.LessThan, Value: 5, Colors: []color.Color{color.Black}}},
					}),
					ContinuousEffects: []game.ContinuousEffect{
						game.ContinuousEffect{
							Layer:          game.LayerType,
							AffectedSource: true,
							RemoveTypes:    []types.Card{types.Creature},
						},
					},
				},
			},
			ActivatedAbilities: []game.ActivatedAbility{
				game.ActivatedAbility{
					Text:     "{1}{B}, Sacrifice another creature: Target creature gets -2/-1 until end of turn.",
					ManaCost: opt.Val(cost.Mana{cost.O(1), cost.B}),
					AdditionalCosts: []cost.Additional{
						{
							Kind:               cost.AdditionalSacrifice,
							Text:               "Sacrifice another creature",
							Amount:             1,
							MatchPermanentType: true,
							PermanentType:      types.Creature,
							ExcludeSource:      true,
						},
					},
					ZoneOfFunction: zone.Battlefield,
					Content: game.Mode{
						Targets: []game.TargetSpec{
							game.TargetSpec{
								MinTargets: 1,
								MaxTargets: 1,
								Constraint: "target creature",
								Allow:      game.TargetAllowPermanent,
								Selection:  opt.Val(game.Selection{RequiredTypesAny: []types.Card{types.Creature}}),
							},
						},
						Sequence: []game.Instruction{
							{
								Primitive: game.ModifyPT{
									Object:         game.TargetPermanentReference(0),
									PowerDelta:     game.Fixed(-2),
									ToughnessDelta: game.Fixed(-1),
									Duration:       game.DurationUntilEndOfTurn,
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
							Controller:       game.TriggerControllerYou,
							ExcludeSelf:      true,
							SubjectSelection: game.Selection{RequiredTypes: []types.Card{types.Creature}},
						},
					},
					Content: game.Mode{
						Sequence: []game.Instruction{
							{
								Primitive: game.LoseLife{
									Amount: game.Fixed(2),
									Player: game.ControllerReference(),
								},
								Optional:      true,
								PublishResult: game.ResultKey("if-you-do"),
							},
							{
								Primitive: game.Draw{
									Amount: game.Fixed(1),
									Player: game.ControllerReference(),
								},
								ResultGate: opt.Val(game.InstructionResultGate{
									Key:       "if-you-do",
									Succeeded: game.TriTrue,
								}),
							},
						},
					}.Ability(),
				},
			},
			OracleText: `
			Indestructible
			As long as your devotion to black is less than five, Erebos isn't a creature.
			Whenever another creature you control dies, you may pay 2 life. If you do, draw a card.
			{1}{B}, Sacrifice another creature: Target creature gets -2/-1 until end of turn.
		`,
		},
	}
}
