package t

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// TeyoTheShieldmage is the card definition for Teyo, the Shieldmage.
//
// Type: Legendary Planeswalker — Teyo
// Cost: {2}{W}
//
// Oracle text:
//
//	You have hexproof. (You can't be the target of spells or abilities your opponents control.)
//	−2: Create a 0/3 white Wall creature token with defender.
var TeyoTheShieldmage = newTeyoTheShieldmage

func newTeyoTheShieldmage() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White),
		CardFace: game.CardFace{
			Name: "Teyo, the Shieldmage",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
				cost.W,
			}),
			Colors:     []color.Color{color.White},
			Supertypes: []types.Super{types.Legendary},
			Types:      []types.Card{types.Planeswalker},
			Subtypes:   []types.Sub{types.Teyo},
			Loyalty:    opt.Val(5),
			StaticAbilities: []game.StaticAbility{
				game.PlayerHexproofStaticBody,
			},
			LoyaltyAbilities: []game.LoyaltyAbility{
				game.LoyaltyAbility{
					LoyaltyCost: -2,
					Content: game.Mode{
						Sequence: []game.Instruction{
							{
								Primitive: game.CreateToken{
									Amount: game.Fixed(1),
									Source: game.TokenDef(teyoTheShieldmageToken),
								},
							},
						},
					}.Ability(),
				},
			},
			OracleText: `
			You have hexproof. (You can't be the target of spells or abilities your opponents control.)
			−2: Create a 0/3 white Wall creature token with defender.
		`,
		},
	}
}

var teyoTheShieldmageToken = newTeyoTheShieldmageToken()

func newTeyoTheShieldmageToken() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name:      "Wall",
			Colors:    []color.Color{color.White},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Wall},
			Power:     opt.Val(game.PT{Value: 0}),
			Toughness: opt.Val(game.PT{Value: 3}),
			StaticAbilities: []game.StaticAbility{
				game.DefenderStaticBody,
			},
		},
	}
}
