package j

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// JaceUnravelerOfSecrets is the card definition for Jace, Unraveler of Secrets.
//
// Type: Legendary Planeswalker — Jace
// Cost: {3}{U}{U}
//
// Oracle text:
//
//	+1: Scry 1, then draw a card.
//	−2: Return target creature to its owner's hand.
//	−8: You get an emblem with "Whenever an opponent casts their first spell each turn, counter that spell."
var JaceUnravelerOfSecrets = newJaceUnravelerOfSecrets

func newJaceUnravelerOfSecrets() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue),
		CardFace: game.CardFace{
			Name: "Jace, Unraveler of Secrets",
			ManaCost: opt.Val(cost.Mana{
				cost.O(3),
				cost.U,
				cost.U,
			}),
			Colors:     []color.Color{color.Blue},
			Supertypes: []types.Super{types.Legendary},
			Types:      []types.Card{types.Planeswalker},
			Subtypes:   []types.Sub{types.Jace},
			Loyalty:    opt.Val(5),
			LoyaltyAbilities: []game.LoyaltyAbility{
				game.LoyaltyAbility{
					LoyaltyCost: 1,
					Content: game.Mode{
						Sequence: []game.Instruction{
							{
								Primitive: game.Scry{
									Amount: game.Fixed(1),
									Player: game.ControllerReference(),
								},
							},
							{
								Primitive: game.Draw{
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
								Primitive: game.Bounce{
									Object: game.TargetPermanentReference(0),
								},
							},
						},
					}.Ability(),
				},
				game.LoyaltyAbility{
					LoyaltyCost: -8,
					Content: game.Mode{
						Sequence: []game.Instruction{
							{
								Primitive: game.CreateEmblem{
									EmblemAbilities: []game.Ability{
										new(game.TriggeredAbility{
											Trigger: game.TriggerCondition{
												Type: game.TriggerWhenever,
												Pattern: game.TriggerPattern{
													Event:                      game.EventSpellCast,
													Controller:                 game.TriggerControllerOpponent,
													PlayerEventOrdinalThisTurn: 1,
												},
											},
											Content: game.Mode{
												Sequence: []game.Instruction{
													{
														Primitive: game.CounterObject{
															Object: game.EventStackObjectReference(),
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
			+1: Scry 1, then draw a card.
			−2: Return target creature to its owner's hand.
			−8: You get an emblem with "Whenever an opponent casts their first spell each turn, counter that spell."
		`,
		},
	}
}
