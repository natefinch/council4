package e

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// ÉomerKingOfRohan is the card definition for Éomer, King of Rohan.
//
// Type: Legendary Creature — Human Noble
// Cost: {3}{R}{W}
//
// Oracle text:
//
//	Double strike
//	Éomer enters with a +1/+1 counter on it for each other Human you control.
//	When Éomer enters, target player becomes the monarch. Éomer deals damage equal to its power to any target.
var ÉomerKingOfRohan = newÉomerKingOfRohan()

func newÉomerKingOfRohan() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White, color.Red),
		CardFace: game.CardFace{
			Name: "Éomer, King of Rohan",
			ManaCost: opt.Val(cost.Mana{
				cost.O(3),
				cost.R,
				cost.W,
			}),
			Colors:     []color.Color{color.Red, color.White},
			Supertypes: []types.Super{types.Legendary},
			Types:      []types.Card{types.Creature},
			Subtypes:   []types.Sub{types.Human, types.Noble},
			Power:      opt.Val(game.PT{Value: 2}),
			Toughness:  opt.Val(game.PT{Value: 2}),
			StaticAbilities: []game.StaticAbility{
				game.DoubleStrikeStaticBody,
			},
			TriggeredAbilities: []game.TriggeredAbility{
				game.TriggeredAbility{
					Trigger: game.TriggerCondition{
						Type: game.TriggerWhen,
						Pattern: game.TriggerPattern{
							Event:  game.EventPermanentEnteredBattlefield,
							Source: game.TriggerSourceSelf,
						},
					},
					Content: game.Mode{
						Targets: []game.TargetSpec{
							game.TargetSpec{
								MinTargets: 1,
								MaxTargets: 1,
								Constraint: "target player",
								Allow:      game.TargetAllowPlayer,
							},
							game.TargetSpec{
								MinTargets: 1,
								MaxTargets: 1,
								Constraint: "any target",
								Allow:      game.TargetAllowPermanent | game.TargetAllowPlayer,
							},
						},
						Sequence: []game.Instruction{
							{
								Primitive: game.BecomeMonarch{
									Player: game.TargetPlayerReference(0),
								},
							},
							{
								Primitive: game.Damage{
									Amount: game.Dynamic(game.DynamicAmount{
										Kind:       game.DynamicAmountObjectPower,
										Multiplier: 1,
										Object:     game.SourcePermanentReference(),
									}),
									Recipient:    game.AnyTargetDamageRecipient(1),
									DamageSource: opt.Val(game.SourcePermanentReference()),
								},
							},
						},
					}.Ability(),
				},
			},
			ReplacementAbilities: []game.ReplacementAbility{
				game.EntersWithCountersReplacement("Éomer enters with a +1/+1 counter on it for each other Human you control.", game.CounterPlacement{Kind: counter.PlusOnePlusOne, Dynamic: opt.Val(&game.DynamicAmount{
					Kind:       game.DynamicAmountCountSelector,
					Multiplier: 1,
					Group:      game.BattlefieldGroup(game.Selection{SubtypesAny: []types.Sub{types.Sub("Human")}, Controller: game.ControllerYou, ExcludeSource: true}),
				})}),
			},
			OracleText: `
			Double strike
			Éomer enters with a +1/+1 counter on it for each other Human you control.
			When Éomer enters, target player becomes the monarch. Éomer deals damage equal to its power to any target.
		`,
		},
	}
}
