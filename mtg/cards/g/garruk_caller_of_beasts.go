package g

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// GarrukCallerOfBeasts is the card definition for Garruk, Caller of Beasts.
//
// Type: Legendary Planeswalker — Garruk
// Cost: {4}{G}{G}
//
// Oracle text:
//
//	+1: Reveal the top five cards of your library. Put all creature cards revealed this way into your hand and the rest on the bottom of your library in any order.
//	−3: You may put a green creature card from your hand onto the battlefield.
//	−7: You get an emblem with "Whenever you cast a creature spell, you may search your library for a creature card, put it onto the battlefield, then shuffle."
var GarrukCallerOfBeasts = newGarrukCallerOfBeasts()

func newGarrukCallerOfBeasts() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Green),
		CardFace: game.CardFace{
			Name: "Garruk, Caller of Beasts",
			ManaCost: opt.Val(cost.Mana{
				cost.O(4),
				cost.G,
				cost.G,
			}),
			Colors:     []color.Color{color.Green},
			Supertypes: []types.Super{types.Legendary},
			Types:      []types.Card{types.Planeswalker},
			Subtypes:   []types.Sub{types.Garruk},
			Loyalty:    opt.Val(4),
			LoyaltyAbilities: []game.LoyaltyAbility{
				game.LoyaltyAbility{
					LoyaltyCost: 1,
					Content: game.Mode{
						Sequence: []game.Instruction{
							{
								Primitive: game.RevealTopPartition{
									Player:    game.ControllerReference(),
									Amount:    game.Fixed(5),
									Selection: game.Selection{RequiredTypes: []types.Card{types.Creature}},
									Remainder: game.DigRemainderLibraryBottom,
								},
							},
						},
					}.Ability(),
				},
				game.LoyaltyAbility{
					LoyaltyCost: -3,
					Content: game.Mode{
						Sequence: []game.Instruction{
							{
								Primitive: game.ChooseFromZone{
									Player:     game.ControllerReference(),
									SourceZone: zone.Hand,
									Filter:     game.Selection{RequiredTypes: []types.Card{types.Creature}, ColorsAny: []color.Color{color.Green}},
									Quantity:   game.Fixed(1),
									Destination: game.ChooseDestination{
										Zone: zone.Battlefield,
									},
									Prompt: "Choose a card to put onto the battlefield",
								},
								Optional: true,
							},
						},
					}.Ability(),
				},
				game.LoyaltyAbility{
					LoyaltyCost: -7,
					Content: game.Mode{
						Sequence: []game.Instruction{
							{
								Primitive: game.CreateEmblem{
									EmblemAbilities: []game.Ability{
										new(game.TriggeredAbility{
											Trigger: game.TriggerCondition{
												Type: game.TriggerWhenever,
												Pattern: game.TriggerPattern{
													Event:         game.EventSpellCast,
													Controller:    game.TriggerControllerYou,
													CardSelection: game.Selection{RequiredTypes: []types.Card{types.Creature}},
												},
											},
											Content: game.Mode{
												Sequence: []game.Instruction{
													{
														Primitive: game.Search{
															Player: game.ControllerReference(),
															Spec: game.SearchSpec{
																SourceZone:  zone.Library,
																Destination: zone.Battlefield,
																Filter:      game.Selection{RequiredTypes: []types.Card{types.Creature}},
															},
															Amount: game.Fixed(1),
														},
														Optional: true,
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
			+1: Reveal the top five cards of your library. Put all creature cards revealed this way into your hand and the rest on the bottom of your library in any order.
			−3: You may put a green creature card from your hand onto the battlefield.
			−7: You get an emblem with "Whenever you cast a creature spell, you may search your library for a creature card, put it onto the battlefield, then shuffle."
		`,
		},
	}
}
