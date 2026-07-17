package s

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// SpellbookVendor is the card definition for Spellbook Vendor.
//
// Type: Creature — Human Peasant
// Cost: {1}{W}
//
// Oracle text:
//
//	Vigilance
//	At the beginning of combat on your turn, you may pay {1}. When you do, create a Sorcerer Role token attached to target creature you control. (If you control another Role on it, put that one into the graveyard. Enchanted creature gets +1/+1 and has "Whenever this creature attacks, scry 1.")
var SpellbookVendor = newSpellbookVendor

func newSpellbookVendor() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White),
		CardFace: game.CardFace{
			Name: "Spellbook Vendor",
			ManaCost: opt.Val(cost.Mana{
				cost.O(1),
				cost.W,
			}),
			Colors:    []color.Color{color.White},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Human, types.Peasant},
			Power:     opt.Val(game.PT{Value: 2}),
			Toughness: opt.Val(game.PT{Value: 2}),
			StaticAbilities: []game.StaticAbility{
				game.VigilanceStaticBody,
			},
			TriggeredAbilities: []game.TriggeredAbility{
				game.TriggeredAbility{
					Trigger: game.TriggerCondition{
						Type: game.TriggerAt,
						Pattern: game.TriggerPattern{
							Event:      game.EventBeginningOfStep,
							Controller: game.TriggerControllerYou,
							Step:       game.StepBeginningOfCombat,
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
														Source:          game.TokenDef(spellbookVendorToken),
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
			Vigilance
			At the beginning of combat on your turn, you may pay {1}. When you do, create a Sorcerer Role token attached to target creature you control. (If you control another Role on it, put that one into the graveyard. Enchanted creature gets +1/+1 and has "Whenever this creature attacks, scry 1.")
		`,
		},
	}
}

var spellbookVendorToken = newSpellbookVendorToken()

func newSpellbookVendorToken() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name:     "Sorcerer Role",
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
									},
									Content: game.Mode{
										Sequence: []game.Instruction{
											{
												Primitive: game.Scry{
													Amount: game.Fixed(1),
													Player: game.ControllerReference(),
												},
											},
										},
									}.Ability(),
								}),
							},
						},
						game.ContinuousEffect{
							Layer:          game.LayerPowerToughnessModify,
							Group:          game.AttachedObjectGroup(game.SourcePermanentReference()),
							PowerDelta:     1,
							ToughnessDelta: 1,
						},
					},
				},
			},
			OracleText: `
			Enchant creature
			Enchanted creature gets +1/+1 and has "Whenever this creature attacks, scry 1."
		`,
		},
	}
}
