package n

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// NissaRevane is the card definition for Nissa Revane.
//
// Type: Legendary Planeswalker — Nissa
// Cost: {2}{G}{G}
//
// Oracle text:
//
//	+1: Search your library for a card named Nissa's Chosen, put it onto the battlefield, then shuffle.
//	+1: You gain 2 life for each Elf you control.
//	−7: Search your library for any number of Elf creature cards, put them onto the battlefield, then shuffle.
var NissaRevane = newNissaRevane

func newNissaRevane() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Green),
		CardFace: game.CardFace{
			Name: "Nissa Revane",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
				cost.G,
				cost.G,
			}),
			Colors:     []color.Color{color.Green},
			Supertypes: []types.Super{types.Legendary},
			Types:      []types.Card{types.Planeswalker},
			Subtypes:   []types.Sub{types.Nissa},
			Loyalty:    opt.Val(2),
			LoyaltyAbilities: []game.LoyaltyAbility{
				game.LoyaltyAbility{
					LoyaltyCost: 1,
					Content: game.Mode{
						Sequence: []game.Instruction{
							{
								Primitive: game.Search{
									Player: game.ControllerReference(),
									Spec: game.SearchSpec{
										SourceZone:  zone.Library,
										Destination: zone.Battlefield,
										Name:        "Nissa's Chosen",
									},
									Amount: game.Fixed(1),
								},
							},
						},
					}.Ability(),
				},
				game.LoyaltyAbility{
					LoyaltyCost: 1,
					Content: game.Mode{
						Sequence: []game.Instruction{
							{
								Primitive: game.GainLife{
									Amount: game.Dynamic(game.DynamicAmount{
										Kind:       game.DynamicAmountCountSelector,
										Multiplier: 2,
										Group:      game.BattlefieldGroup(game.Selection{SubtypesAny: []types.Sub{types.Sub("Elf")}, Controller: game.ControllerYou}),
									}),
									Player: game.ControllerReference(),
								},
							},
						},
					}.Ability(),
				},
				game.LoyaltyAbility{
					LoyaltyCost: -7,
					Content: game.Mode{
						Sequence: []game.Instruction{
							{
								Primitive: game.Search{
									Player: game.ControllerReference(),
									Spec: game.SearchSpec{
										SourceZone:  zone.Library,
										Destination: zone.Battlefield,
										Filter:      game.Selection{RequiredTypes: []types.Card{types.Creature}, SubtypesAny: []types.Sub{types.Sub("Elf")}},
										AnyNumber:   true,
									},
									Amount: game.Fixed(0),
								},
							},
						},
					}.Ability(),
				},
			},
			OracleText: `
			+1: Search your library for a card named Nissa's Chosen, put it onto the battlefield, then shuffle.
			+1: You gain 2 life for each Elf you control.
			−7: Search your library for any number of Elf creature cards, put them onto the battlefield, then shuffle.
		`,
		},
	}
}
