package a

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// ArceeSharpshooter is the card definition for Arcee, Sharpshooter // Arcee, Acrobatic Coupe.
//
// Type: Legendary Artifact Creature — Robot // Legendary Artifact — Vehicle
// Face: Arcee, Acrobatic Coupe — Legendary Artifact — Vehicle
//
// Oracle text:
//
//	More Than Meets the Eye {R}{W} (You may cast this card converted for {R}{W}.)
//	First strike
//	{1}, Remove one or more +1/+1 counters from Arcee: It deals that much damage to target creature. Convert Arcee.
var ArceeSharpshooter = newArceeSharpshooter

func newArceeSharpshooter() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White, color.Red),
		CardFace: game.CardFace{
			Name: "Arcee, Sharpshooter",
			ManaCost: opt.Val(cost.Mana{
				cost.O(1),
				cost.R,
				cost.W,
			}),
			Colors:     []color.Color{color.Red, color.White},
			Supertypes: []types.Super{types.Legendary},
			Types:      []types.Card{types.Artifact, types.Creature},
			Subtypes:   []types.Sub{types.Robot},
			Power:      opt.Val(game.PT{Value: 2}),
			Toughness:  opt.Val(game.PT{Value: 2}),
			StaticAbilities: []game.StaticAbility{
				game.FirstStrikeStaticBody,
			},
			ActivatedAbilities: []game.ActivatedAbility{
				game.ActivatedAbility{
					Text:     "{1}, Remove one or more +1/+1 counters from Arcee: It deals that much damage to target creature. Convert Arcee.",
					ManaCost: opt.Val(cost.Mana{cost.O(1)}),
					AdditionalCosts: []cost.Additional{
						{
							Kind:             cost.AdditionalRemoveCounter,
							Text:             "Remove one or more +1/+1 counters from Arcee",
							AmountFromX:      true,
							AmountAtLeastOne: true,
							CounterKind:      counter.PlusOnePlusOne,
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
								Primitive: game.Damage{
									Amount: game.Dynamic(game.DynamicAmount{
										Kind:       game.DynamicAmountX,
										Multiplier: 1,
									}),
									Recipient:    game.AnyTargetDamageRecipient(0),
									DamageSource: opt.Val(game.SourcePermanentReference()),
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
					ManaCost: opt.Val(cost.Mana{cost.R, cost.W}),
					Mechanic: cost.AlternativeMechanicMoreThanMeetsTheEye,
				},
			},
			OracleText: `
			More Than Meets the Eye {R}{W} (You may cast this card converted for {R}{W}.)
			First strike
			{1}, Remove one or more +1/+1 counters from Arcee: It deals that much damage to target creature. Convert Arcee.
		`,
		},
		Layout: game.LayoutTransform,
		Back: opt.Val(game.CardFace{
			Name:       "Arcee, Acrobatic Coupe",
			Colors:     []color.Color{color.Red, color.White},
			Supertypes: []types.Super{types.Legendary},
			Types:      []types.Card{types.Artifact},
			Subtypes:   []types.Sub{types.Vehicle},
			Power:      opt.Val(game.PT{Value: 2}),
			Toughness:  opt.Val(game.PT{Value: 2}),
			StaticAbilities: []game.StaticAbility{
				game.LivingMetalStaticBody,
			},
			TriggeredAbilities: []game.TriggeredAbility{
				game.TriggeredAbility{
					Trigger: game.TriggerCondition{
						Type: game.TriggerWhenever,
						Pattern: game.TriggerPattern{
							Event:              game.EventSpellCast,
							Controller:         game.TriggerControllerYou,
							SpellTargetAllow:   game.TargetAllowPermanent,
							SpellTargetPattern: opt.Val(game.Selection{AnyOf: []game.Selection{game.Selection{RequiredTypesAny: []types.Card{types.Creature}}, game.Selection{SubtypesAny: []types.Sub{types.Sub("Vehicle")}}}, Controller: game.ControllerYou}),
						},
					},
					Content: game.Mode{
						Sequence: []game.Instruction{
							{
								Primitive: game.AddCounter{
									Amount: game.Dynamic(game.DynamicAmount{
										Kind:       game.DynamicAmountSpellTargetCount,
										Multiplier: 1,
										Selection:  &game.Selection{AnyOf: []game.Selection{game.Selection{RequiredTypesAny: []types.Card{types.Creature}}, game.Selection{SubtypesAny: []types.Sub{types.Sub("Vehicle")}}}, Controller: game.ControllerYou},
									}),
									Object:      game.SourcePermanentReference(),
									CounterKind: counter.PlusOnePlusOne,
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
			OracleText: `
			Living metal (During your turn, this Vehicle is also a creature.)
			Whenever you cast a spell that targets one or more creatures or Vehicles you control, put that many +1/+1 counters on Arcee. Convert Arcee.
		`,
		}),
	}
}
