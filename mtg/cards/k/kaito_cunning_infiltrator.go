package k

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// KaitoCunningInfiltrator is the card definition for Kaito, Cunning Infiltrator.
//
// Type: Legendary Planeswalker — Kaito
// Cost: {1}{U}{U}
//
// Oracle text:
//
//	Whenever a creature you control deals combat damage to a player, put a loyalty counter on Kaito.
//	+1: Up to one target creature you control can't be blocked this turn. Draw a card, then discard a card.
//	−2: Create a 2/1 blue Ninja creature token.
//	−9: You get an emblem with "Whenever a player casts a spell, you create a 2/1 blue Ninja creature token."
var KaitoCunningInfiltrator = newKaitoCunningInfiltrator()

func newKaitoCunningInfiltrator() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue),
		CardFace: game.CardFace{
			Name: "Kaito, Cunning Infiltrator",
			ManaCost: opt.Val(cost.Mana{
				cost.O(1),
				cost.U,
				cost.U,
			}),
			Colors:     []color.Color{color.Blue},
			Supertypes: []types.Super{types.Legendary},
			Types:      []types.Card{types.Planeswalker},
			Subtypes:   []types.Sub{types.Kaito},
			Loyalty:    opt.Val(3),
			TriggeredAbilities: []game.TriggeredAbility{
				game.TriggeredAbility{
					Trigger: game.TriggerCondition{
						Type: game.TriggerWhenever,
						Pattern: game.TriggerPattern{
							Event:                 game.EventDamageDealt,
							Controller:            game.TriggerControllerYou,
							Subject:               game.TriggerSubjectDamageSource,
							RequireCombatDamage:   true,
							DamageRecipient:       game.DamageRecipientPlayer,
							DamageSourceSelection: game.Selection{RequiredTypes: []types.Card{types.Creature}},
						},
					},
					Content: game.Mode{
						Sequence: []game.Instruction{
							{
								Primitive: game.AddCounter{
									Amount:      game.Fixed(1),
									Object:      game.SourcePermanentReference(),
									CounterKind: counter.Loyalty,
								},
							},
						},
					}.Ability(),
				},
			},
			LoyaltyAbilities: []game.LoyaltyAbility{
				game.LoyaltyAbility{
					LoyaltyCost: 1,
					Content: game.Mode{
						Targets: []game.TargetSpec{
							game.TargetSpec{
								MinTargets: 0,
								MaxTargets: 1,
								Constraint: "up to one target creature you control",
								Allow:      game.TargetAllowPermanent,
								Selection:  opt.Val(game.Selection{RequiredTypesAny: []types.Card{types.Creature}, Controller: game.ControllerYou}),
							},
						},
						Sequence: []game.Instruction{
							{
								Primitive: game.ApplyRule{
									Object: opt.Val(game.TargetPermanentReference(0)),
									RuleEffects: []game.RuleEffect{
										game.RuleEffect{
											Kind: game.RuleEffectCantBeBlocked,
										},
									},
									Duration: game.DurationThisTurn,
								},
							},
							{
								Primitive: game.Draw{
									Amount: game.Fixed(1),
									Player: game.ControllerReference(),
								},
							},
							{
								Primitive: game.Discard{
									Amount: game.Fixed(1),
									Player: game.ControllerReference(),
								},
							},
						},
					}.Ability(),
				},
				game.LoyaltyAbility{
					LoyaltyCost: -2,
					Content: game.Mode{
						Sequence: []game.Instruction{
							{
								Primitive: game.CreateToken{
									Amount: game.Fixed(1),
									Source: game.TokenDef(kaitoCunningInfiltratorToken),
								},
							},
						},
					}.Ability(),
				},
				game.LoyaltyAbility{
					LoyaltyCost: -9,
					Content: game.Mode{
						Sequence: []game.Instruction{
							{
								Primitive: game.CreateEmblem{
									EmblemAbilities: []game.Ability{
										new(game.TriggeredAbility{
											Trigger: game.TriggerCondition{
												Type: game.TriggerWhenever,
												Pattern: game.TriggerPattern{
													Event: game.EventSpellCast,
												},
											},
											Content: game.Mode{
												Sequence: []game.Instruction{
													{
														Primitive: game.CreateToken{
															Amount: game.Fixed(1),
															Source: game.TokenDef(kaitoCunningInfiltratorToken),
														},
													},
												},
											}.Ability(),
										}),
									},
								},
							},
						},
					}.Ability(),
				},
			},
			OracleText: `
			Whenever a creature you control deals combat damage to a player, put a loyalty counter on Kaito.
			+1: Up to one target creature you control can't be blocked this turn. Draw a card, then discard a card.
			−2: Create a 2/1 blue Ninja creature token.
			−9: You get an emblem with "Whenever a player casts a spell, you create a 2/1 blue Ninja creature token."
		`,
		},
	}
}

var kaitoCunningInfiltratorToken = newKaitoCunningInfiltratorToken()

func newKaitoCunningInfiltratorToken() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name:      "Ninja",
			Colors:    []color.Color{color.Blue},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Ninja},
			Power:     opt.Val(game.PT{Value: 2}),
			Toughness: opt.Val(game.PT{Value: 1}),
		},
	}
}
