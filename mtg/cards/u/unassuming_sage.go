package u

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// UnassumingSage is the card definition for Unassuming Sage.
//
// Type: Creature — Human Peasant Wizard
// Cost: {1}{W}
//
// Oracle text:
//
//	When this creature enters, you may pay {2}. If you do, create a Sorcerer Role token attached to it. (Enchanted creature gets +1/+1 and has "Whenever this creature attacks, scry 1.")
var UnassumingSage = newUnassumingSage

func newUnassumingSage() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White),
		CardFace: game.CardFace{
			Name: "Unassuming Sage",
			ManaCost: opt.Val(cost.Mana{
				cost.O(1),
				cost.W,
			}),
			Colors:    []color.Color{color.White},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Human, types.Peasant, types.Wizard},
			Power:     opt.Val(game.PT{Value: 2}),
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
										Prompt: "Pay {2}?",
										ManaCost: opt.Val(cost.Mana{
											cost.O(2),
										}),
									},
								},
								PublishResult: game.ResultKey("controller-paid"),
							},
							{
								Primitive: game.CreateToken{
									Amount:          game.Fixed(1),
									Source:          game.TokenDef(unassumingSageToken),
									EntryAttachedTo: opt.Val(game.EventPermanentReference()),
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
			When this creature enters, you may pay {2}. If you do, create a Sorcerer Role token attached to it. (Enchanted creature gets +1/+1 and has "Whenever this creature attacks, scry 1.")
		`,
		},
	}
}

var unassumingSageToken = newUnassumingSageToken()

func newUnassumingSageToken() *game.CardDef {
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
