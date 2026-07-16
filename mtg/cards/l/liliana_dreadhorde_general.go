package l

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// LilianaDreadhordeGeneral is the card definition for Liliana, Dreadhorde General.
//
// Type: Legendary Planeswalker — Liliana
// Cost: {4}{B}{B}
//
// Oracle text:
//
//	Whenever a creature you control dies, draw a card.
//	+1: Create a 2/2 black Zombie creature token.
//	−4: Each player sacrifices two creatures of their choice.
//	−9: Each opponent chooses a permanent they control of each permanent type and sacrifices the rest.
var LilianaDreadhordeGeneral = newLilianaDreadhordeGeneral

func newLilianaDreadhordeGeneral() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Black),
		CardFace: game.CardFace{
			Name: "Liliana, Dreadhorde General",
			ManaCost: opt.Val(cost.Mana{
				cost.O(4),
				cost.B,
				cost.B,
			}),
			Colors:     []color.Color{color.Black},
			Supertypes: []types.Super{types.Legendary},
			Types:      []types.Card{types.Planeswalker},
			Subtypes:   []types.Sub{types.Liliana},
			Loyalty:    opt.Val(6),
			TriggeredAbilities: []game.TriggeredAbility{
				game.TriggeredAbility{
					Trigger: game.TriggerCondition{
						Type: game.TriggerWhenever,
						Pattern: game.TriggerPattern{
							Event:            game.EventPermanentDied,
							Controller:       game.TriggerControllerYou,
							SubjectSelection: game.Selection{RequiredTypes: []types.Card{types.Creature}},
						},
					},
					Content: game.Mode{
						Sequence: []game.Instruction{
							{
								Primitive: game.Draw{
									Amount: game.Fixed(1),
									Player: game.ControllerReference(),
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
						Sequence: []game.Instruction{
							{
								Primitive: game.CreateToken{
									Amount: game.Fixed(1),
									Source: game.TokenDef(lilianaDreadhordeGeneralToken),
								},
							},
						},
					}.Ability(),
				},
				game.LoyaltyAbility{
					LoyaltyCost: -4,
					Content: game.Mode{
						Sequence: []game.Instruction{
							{
								Primitive: game.SacrificePermanents{
									Amount:      game.Fixed(2),
									PlayerGroup: game.AllPlayersReference(),
									Selection:   game.Selection{RequiredTypes: []types.Card{types.Creature}},
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
								Primitive: game.KeepOnePerType{
									Players: game.OpponentsReference(),
									Types:   []types.Card{types.Artifact, types.Battle, types.Creature, types.Enchantment, types.Land, types.Planeswalker},
								},
							},
						},
					}.Ability(),
				},
			},
			OracleText: `
			Whenever a creature you control dies, draw a card.
			+1: Create a 2/2 black Zombie creature token.
			−4: Each player sacrifices two creatures of their choice.
			−9: Each opponent chooses a permanent they control of each permanent type and sacrifices the rest.
		`,
		},
	}
}

var lilianaDreadhordeGeneralToken = newLilianaDreadhordeGeneralToken()

func newLilianaDreadhordeGeneralToken() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name:      "Zombie",
			Colors:    []color.Color{color.Black},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Zombie},
			Power:     opt.Val(game.PT{Value: 2}),
			Toughness: opt.Val(game.PT{Value: 2}),
		},
	}
}
