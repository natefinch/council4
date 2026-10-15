package x

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// XenagosTheReveler is the card definition for Xenagos, the Reveler.
//
// Type: Legendary Planeswalker — Xenagos
// Cost: {2}{R}{G}
//
// Oracle text:
//
//	+1: Add X mana in any combination of {R} and/or {G}, where X is the number of creatures you control.
//	0: Create a 2/2 red and green Satyr creature token with haste.
//	−6: Exile the top seven cards of your library. You may put any number of creature and/or land cards from among them onto the battlefield.
var XenagosTheReveler = newXenagosTheReveler

func newXenagosTheReveler() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Red, color.Green),
		CardFace: game.CardFace{
			Name: "Xenagos, the Reveler",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
				cost.R,
				cost.G,
			}),
			Colors:     []color.Color{color.Green, color.Red},
			Supertypes: []types.Super{types.Legendary},
			Types:      []types.Card{types.Planeswalker},
			Subtypes:   []types.Sub{types.Xenagos},
			Loyalty:    opt.Val(3),
			LoyaltyAbilities: []game.LoyaltyAbility{
				game.LoyaltyAbility{
					Text:        "+1: Add X mana in any combination of {R} and/or {G}, where X is the number of creatures you control.",
					LoyaltyCost: 1,
					Content: game.AbilityContent{
						Modes: []game.Mode{
							game.Mode{
								Sequence: []game.Instruction{
									game.Instruction{
										Primitive: game.AddMana{
											Amount: game.Dynamic(game.DynamicAmount{
												Kind:       game.DynamicAmountCountSelector,
												Multiplier: 1,
												Group:      game.BattlefieldGroup(game.Selection{RequiredTypes: []types.Card{types.Creature}, Controller: game.ControllerYou}),
											}),
											CombinationColors: []mana.Color{mana.R, mana.G},
										},
									},
								},
							},
						},
						MinModes: 1,
						MaxModes: 1,
					},
				},
				game.LoyaltyAbility{
					Text: "0: Create a 2/2 red and green Satyr creature token with haste.",
					Content: game.AbilityContent{
						Modes: []game.Mode{
							game.Mode{
								Sequence: []game.Instruction{
									game.Instruction{
										Primitive: game.CreateToken{
											Amount: game.Fixed(1),
											Source: game.TokenDef(xenagosTheRevelerToken),
										},
									},
								},
							},
						},
						MinModes: 1,
						MaxModes: 1,
					},
				},
				game.LoyaltyAbility{
					Text:        "−6: Exile the top seven cards of your library. You may put any number of creature and/or land cards from among them onto the battlefield.",
					LoyaltyCost: -6,
					Content: game.AbilityContent{
						Modes: []game.Mode{
							game.Mode{
								Sequence: []game.Instruction{
									game.Instruction{
										Primitive: game.MoveTopOfLibrary{
											Amount:        game.Fixed(7),
											Player:        game.ControllerReference(),
											Destination:   zone.Exile,
											PublishLinked: game.LinkedKey("exiled-top-cards"),
										},
									},
									game.Instruction{
										Primitive: game.ChooseFromZone{
											Player:     game.ControllerReference(),
											SourceZone: zone.Exile,
											Filter:     game.Selection{RequiredTypesAny: []types.Card{types.Creature, types.Land}},
											Count:      game.ChooseAnyNumber,
											Destination: game.ChooseDestination{
												Zone: zone.Battlefield,
											},
											Riders: game.ChooseRiders{
												FromLinked: game.LinkedKey("exiled-top-cards"),
											},
										},
									},
								},
							},
						},
						MinModes: 1,
						MaxModes: 1,
					},
				},
			},
			OracleText: `
			+1: Add X mana in any combination of {R} and/or {G}, where X is the number of creatures you control.
			0: Create a 2/2 red and green Satyr creature token with haste.
			−6: Exile the top seven cards of your library. You may put any number of creature and/or land cards from among them onto the battlefield.
		`,
		},
	}
}

var xenagosTheRevelerToken = newXenagosTheRevelerToken()

func newXenagosTheRevelerToken() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name:      "Satyr",
			Colors:    []color.Color{color.Red, color.Green},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Satyr},
			Power:     opt.Val(game.PT{Value: 2}),
			Toughness: opt.Val(game.PT{Value: 2}),
			StaticAbilities: []game.StaticAbility{
				game.HasteStaticBody,
			},
		},
	}
}
