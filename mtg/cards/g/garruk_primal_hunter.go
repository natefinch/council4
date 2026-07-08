package g

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// GarrukPrimalHunter is the card definition for Garruk, Primal Hunter.
//
// Type: Legendary Planeswalker — Garruk
// Cost: {2}{G}{G}{G}
//
// Oracle text:
//
//	+1: Create a 3/3 green Beast creature token.
//	−3: Draw cards equal to the greatest power among creatures you control.
//	−6: Create a 6/6 green Wurm creature token for each land you control.
var GarrukPrimalHunter = newGarrukPrimalHunter

func newGarrukPrimalHunter() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Green),
		CardFace: game.CardFace{
			Name: "Garruk, Primal Hunter",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
				cost.G,
				cost.G,
				cost.G,
			}),
			Colors:     []color.Color{color.Green},
			Supertypes: []types.Super{types.Legendary},
			Types:      []types.Card{types.Planeswalker},
			Subtypes:   []types.Sub{types.Garruk},
			Loyalty:    opt.Val(3),
			LoyaltyAbilities: []game.LoyaltyAbility{
				game.LoyaltyAbility{
					LoyaltyCost: 1,
					Content: game.Mode{
						Sequence: []game.Instruction{
							{
								Primitive: game.CreateToken{
									Amount: game.Fixed(1),
									Source: game.TokenDef(garrukPrimalHunterToken),
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
								Primitive: game.Draw{
									Amount: game.Dynamic(game.DynamicAmount{
										Kind:       game.DynamicAmountGreatestPowerInGroup,
										Multiplier: 1,
										Group:      game.BattlefieldGroup(game.Selection{RequiredTypes: []types.Card{types.Creature}, Controller: game.ControllerYou}),
									}),
									Player: game.ControllerReference(),
								},
							},
						},
					}.Ability(),
				},
				game.LoyaltyAbility{
					LoyaltyCost: -6,
					Content: game.Mode{
						Sequence: []game.Instruction{
							{
								Primitive: game.CreateToken{
									Amount: game.Dynamic(game.DynamicAmount{
										Kind:       game.DynamicAmountCountSelector,
										Multiplier: 1,
										Group:      game.BattlefieldGroup(game.Selection{RequiredTypes: []types.Card{types.Land}, Controller: game.ControllerYou}),
									}),
									Source: game.TokenDef(garrukPrimalHunterToken2),
								},
							},
						},
					}.Ability(),
				},
			},
			OracleText: `
			+1: Create a 3/3 green Beast creature token.
			−3: Draw cards equal to the greatest power among creatures you control.
			−6: Create a 6/6 green Wurm creature token for each land you control.
		`,
		},
	}
}

var garrukPrimalHunterToken = newGarrukPrimalHunterToken()

func newGarrukPrimalHunterToken() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name:      "Beast",
			Colors:    []color.Color{color.Green},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Beast},
			Power:     opt.Val(game.PT{Value: 3}),
			Toughness: opt.Val(game.PT{Value: 3}),
		},
	}
}

var garrukPrimalHunterToken2 = newGarrukPrimalHunterToken2()

func newGarrukPrimalHunterToken2() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name:      "Wurm",
			Colors:    []color.Color{color.Green},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Wurm},
			Power:     opt.Val(game.PT{Value: 6}),
			Toughness: opt.Val(game.PT{Value: 6}),
		},
	}
}
