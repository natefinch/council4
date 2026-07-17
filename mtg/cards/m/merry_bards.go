package m

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/compare"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// MerryBards is the card definition for Merry Bards.
//
// Type: Creature — Human Bard
// Cost: {2}{R}
//
// Oracle text:
//
//	When this creature enters, you may pay {1}. When you do, create a Young Hero Role token attached to target creature you control. (If you control another Role on it, put that one into the graveyard. Enchanted creature has "Whenever this creature attacks, if its toughness is 3 or less, put a +1/+1 counter on it.")
var MerryBards = newMerryBards

func newMerryBards() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Red),
		CardFace: game.CardFace{
			Name: "Merry Bards",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
				cost.R,
			}),
			Colors:    []color.Color{color.Red},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Human, types.Bard},
			Power:     opt.Val(game.PT{Value: 3}),
			Toughness: opt.Val(game.PT{Value: 2}),
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
						Sequence: []game.Instruction{
							{
								Primitive: game.Pay{
									Payment: game.ResolutionPayment{
										Prompt: "Pay {1}?",
										ManaCost: opt.Val(cost.Mana{
											cost.O(1),
										}),
									},
								},
								PublishResult: game.ResultKey("controller-paid"),
							},
							{
								Primitive: game.CreateReflexiveTrigger{
									Trigger: game.ReflexiveTriggerDef{
										Content: game.Mode{
											Targets: []game.TargetSpec{
												game.TargetSpec{
													MinTargets: 1,
													MaxTargets: 1,
													Constraint: "target creature you control",
													Allow:      game.TargetAllowPermanent,
													Selection:  opt.Val(game.Selection{RequiredTypesAny: []types.Card{types.Creature}, Controller: game.ControllerYou}),
												},
											},
											Sequence: []game.Instruction{
												{
													Primitive: game.CreateToken{
														Amount:          game.Fixed(1),
														Source:          game.TokenDef(merryBardsToken),
														EntryAttachedTo: opt.Val(game.TargetObjectReference(0)),
													},
												},
											},
										}.Ability(),
									},
								},
								ResultGate: opt.Val(game.InstructionResultGate{
									Key:       "controller-paid",
									Succeeded: game.TriTrue,
								}),
							},
						},
					}.Ability(),
				},
			},
			OracleText: `
			When this creature enters, you may pay {1}. When you do, create a Young Hero Role token attached to target creature you control. (If you control another Role on it, put that one into the graveyard. Enchanted creature has "Whenever this creature attacks, if its toughness is 3 or less, put a +1/+1 counter on it.")
		`,
		},
	}
}

var merryBardsToken = newMerryBardsToken()

func newMerryBardsToken() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name:     "Young Hero Role",
			Types:    []types.Card{types.Enchantment},
			Subtypes: []types.Sub{types.Aura, types.Role},
			StaticAbilities: []game.StaticAbility{
				game.EnchantStaticAbility(&game.TargetSpec{
					MinTargets: 1,
					MaxTargets: 1,
					Constraint: "creature",
					Allow:      game.TargetAllowPermanent,
					Selection:  opt.Val(game.Selection{RequiredTypesAny: []types.Card{types.Creature}}),
				}),
				game.StaticAbility{
					ContinuousEffects: []game.ContinuousEffect{
						game.ContinuousEffect{
							Layer: game.LayerAbility,
							Group: game.AttachedObjectGroup(game.SourcePermanentReference()),
							AddAbilities: []game.Ability{
								new(game.TriggeredAbility{
									Trigger: game.TriggerCondition{
										Type: game.TriggerWhenever,
										Pattern: game.TriggerPattern{
											Event:  game.EventAttackerDeclared,
											Source: game.TriggerSourceSelf,
										},
										InterveningIf: "if its toughness is 3 or less",
										InterveningCondition: opt.Val(game.Condition{
											Object:        opt.Val(game.EventPermanentReference()),
											ObjectMatches: opt.Val(game.Selection{Toughness: opt.Val(compare.Int{Op: compare.LessOrEqual, Value: 3})}),
										}),
									},
									Content: game.Mode{
										Sequence: []game.Instruction{
											{
												Primitive: game.AddCounter{
													Amount:      game.Fixed(1),
													Object:      game.EventPermanentReference(),
													CounterKind: counter.PlusOnePlusOne,
												},
											},
										},
									}.Ability(),
								}),
							},
						},
					},
				},
			},
			OracleText: `
			Enchant creature
			Enchanted creature has "Whenever this creature attacks, if its toughness is 3 or less, put a +1/+1 counter on it."
		`,
		},
	}
}
