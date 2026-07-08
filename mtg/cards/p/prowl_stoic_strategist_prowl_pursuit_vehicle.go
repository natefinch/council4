package p

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// ProwlStoicStrategist is the card definition for Prowl, Stoic Strategist // Prowl, Pursuit Vehicle.
//
// Type: Legendary Artifact Creature — Robot // Legendary Artifact — Vehicle
// Face: Prowl, Pursuit Vehicle — Legendary Artifact — Vehicle
//
// Oracle text:
//
//	More Than Meets the Eye {2}{W} (You may cast this card converted for {2}{W}.)
//	Whenever Prowl attacks, exile up to one other target tapped creature or Vehicle. For as long as that card remains exiled, its owner may play it.
//	Whenever a player plays a card exiled with Prowl, you draw a card and convert Prowl.
var ProwlStoicStrategist = newProwlStoicStrategist

func newProwlStoicStrategist() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White),
		CardFace: game.CardFace{
			Name: "Prowl, Stoic Strategist",
			ManaCost: opt.Val(cost.Mana{
				cost.O(3),
				cost.W,
			}),
			Colors:     []color.Color{color.White},
			Supertypes: []types.Super{types.Legendary},
			Types:      []types.Card{types.Artifact, types.Creature},
			Subtypes:   []types.Sub{types.Robot},
			Power:      opt.Val(game.PT{Value: 3}),
			Toughness:  opt.Val(game.PT{Value: 3}),
			TriggeredAbilities: []game.TriggeredAbility{
				game.TriggeredAbility{
					Trigger: game.TriggerCondition{
						Type: game.TriggerWhenever,
						Pattern: game.TriggerPattern{
							Event:  game.EventAttackerDeclared,
							Source: game.TriggerSourceSelf,
						},
					},
					Content: game.Mode{
						Targets: []game.TargetSpec{
							game.TargetSpec{
								MinTargets: 0,
								MaxTargets: 1,
								Allow:      game.TargetAllowPermanent,
								Selection:  opt.Val(game.Selection{AnyOf: []game.Selection{game.Selection{RequiredTypesAny: []types.Card{types.Creature}}, game.Selection{SubtypesAny: []types.Sub{types.Sub("Vehicle")}}}, Tapped: game.TriTrue, ExcludeSource: true}),
							},
						},
						Sequence: []game.Instruction{
							{
								Primitive: game.ExilePermanentForPlay{
									Object:    game.TargetPermanentReference(0),
									LinkedKey: game.LinkedKey("exiled-with-source"),
								},
							},
						},
					}.Ability(),
				},
				game.TriggeredAbility{
					Trigger: game.TriggerCondition{
						Type: game.TriggerWhenever,
						Pattern: game.TriggerPattern{
							Event:                game.EventCardPlayedFromExile,
							PlaysLinkedExileCard: game.LinkedKey("exiled-with-source"),
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
							{
								Primitive: game.Transform{
									Object: game.SourcePermanentReference(),
								},
							},
						},
					}.Ability(),
				},
			},
			AlternativeCosts: []cost.Alternative{
				cost.Alternative{
					Label:    "More Than Meets the Eye",
					ManaCost: opt.Val(cost.Mana{cost.O(2), cost.W}),
					Mechanic: cost.AlternativeMechanicMoreThanMeetsTheEye,
				},
			},
			OracleText: `
			More Than Meets the Eye {2}{W} (You may cast this card converted for {2}{W}.)
			Whenever Prowl attacks, exile up to one other target tapped creature or Vehicle. For as long as that card remains exiled, its owner may play it.
			Whenever a player plays a card exiled with Prowl, you draw a card and convert Prowl.
		`,
		},
		Layout: game.LayoutTransform,
		Back: opt.Val(game.CardFace{
			Name:       "Prowl, Pursuit Vehicle",
			Colors:     []color.Color{color.White},
			Supertypes: []types.Super{types.Legendary},
			Types:      []types.Card{types.Artifact},
			Subtypes:   []types.Sub{types.Vehicle},
			Power:      opt.Val(game.PT{Value: 2}),
			Toughness:  opt.Val(game.PT{Value: 3}),
			StaticAbilities: []game.StaticAbility{
				game.LivingMetalStaticBody,
			},
			TriggeredAbilities: []game.TriggeredAbility{
				game.TriggeredAbility{
					Trigger: game.TriggerCondition{
						Type: game.TriggerWhenever,
						Pattern: game.TriggerPattern{
							Event:            game.EventPermanentEnteredBattlefield,
							Controller:       game.TriggerControllerYou,
							ExcludeSelf:      true,
							SubjectSelection: game.Selection{AnyOf: []game.Selection{game.Selection{RequiredTypesAny: []types.Card{types.Creature}}, game.Selection{SubtypesAny: []types.Sub{types.Sub("Vehicle")}}}},
						},
					},
					CountsResolutionsThisTurn: true,
					Content: game.Mode{
						Sequence: []game.Instruction{
							{
								Primitive: game.AddCounter{
									Amount:      game.Fixed(1),
									Object:      game.SourcePermanentReference(),
									CounterKind: counter.PlusOnePlusOne,
								},
							},
							{
								Primitive: game.Transform{
									Object: game.SourcePermanentReference(),
								},
								Condition: opt.Val(game.EffectCondition{
									Condition: opt.Val(game.Condition{
										SourceAbilityResolutionOrdinalThisTurn: 2,
									}),
								}),
							},
						},
					}.Ability(),
				},
			},
			OracleText: `
			Living metal (During your turn, this Vehicle is also a creature.)
			Whenever another creature or Vehicle you control enters, put a +1/+1 counter on Prowl. If this is the second time this ability has resolved this turn, convert Prowl.
		`,
		}),
	}
}
