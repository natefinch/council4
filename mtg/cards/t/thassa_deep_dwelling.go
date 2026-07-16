package t

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/compare"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// ThassaDeepDwelling is the card definition for Thassa, Deep-Dwelling.
//
// Type: Legendary Enchantment Creature — God
// Cost: {3}{U}
//
// Oracle text:
//
//	Indestructible
//	As long as your devotion to blue is less than five, Thassa isn't a creature.
//	At the beginning of your end step, exile up to one other target creature you control, then return that card to the battlefield under your control.
//	{3}{U}: Tap another target creature.
var ThassaDeepDwelling = newThassaDeepDwelling

func newThassaDeepDwelling() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue),
		CardFace: game.CardFace{
			Name: "Thassa, Deep-Dwelling",
			ManaCost: opt.Val(cost.Mana{
				cost.O(3),
				cost.U,
			}),
			Colors:     []color.Color{color.Blue},
			Supertypes: []types.Super{types.Legendary},
			Types:      []types.Card{types.Enchantment, types.Creature},
			Subtypes:   []types.Sub{types.God},
			Power:      opt.Val(game.PT{Value: 6}),
			Toughness:  opt.Val(game.PT{Value: 5}),
			StaticAbilities: []game.StaticAbility{
				game.IndestructibleStaticBody,
				game.StaticAbility{
					Condition: opt.Val(game.Condition{
						Aggregates: []game.AggregateComparison{{Aggregate: game.AggregateControllerDevotion, Op: compare.LessThan, Value: 5, Colors: []color.Color{color.Blue}}},
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
					Text:           "{3}{U}: Tap another target creature.",
					ManaCost:       opt.Val(cost.Mana{cost.O(3), cost.U}),
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
								Primitive: game.Tap{
									Object: game.TargetPermanentReference(0),
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
							Step:       game.StepEnd,
						},
					},
					Content: game.Mode{
						Targets: []game.TargetSpec{
							game.TargetSpec{
								MinTargets: 0,
								MaxTargets: 1,
								Constraint: "up to one other target creature you control",
								Allow:      game.TargetAllowPermanent,
								Selection:  opt.Val(game.Selection{RequiredTypesAny: []types.Card{types.Creature}, Controller: game.ControllerYou, ExcludeSource: true}),
							},
						},
						Sequence: []game.Instruction{
							{
								Primitive: game.Exile{
									Object:         game.TargetPermanentReference(0),
									ExileLinkedKey: game.LinkedKey("blink-1"),
								},
							},
							{
								Primitive: game.PutOnBattlefield{
									Source:    game.LinkedBattlefieldSource(game.LinkedKey("blink-1")),
									Recipient: opt.Val(game.ControllerReference()),
								},
							},
						},
					}.Ability(),
				},
			},
			OracleText: `
			Indestructible
			As long as your devotion to blue is less than five, Thassa isn't a creature.
			At the beginning of your end step, exile up to one other target creature you control, then return that card to the battlefield under your control.
			{3}{U}: Tap another target creature.
		`,
		},
	}
}
