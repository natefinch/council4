package a

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// AyulaSInfluence is the card definition for Ayula's Influence.
//
// Type: Enchantment
// Cost: {G}{G}{G}
//
// Oracle text:
//
//	Discard a land card: Create a 2/2 green Bear creature token.
var AyulaSInfluence = newAyulaSInfluence

func newAyulaSInfluence() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Green),
		CardFace: game.CardFace{
			Name: "Ayula's Influence",
			ManaCost: opt.Val(cost.Mana{
				cost.G,
				cost.G,
				cost.G,
			}),
			Colors: []color.Color{color.Green},
			Types:  []types.Card{types.Enchantment},
			ActivatedAbilities: []game.ActivatedAbility{
				game.ActivatedAbility{
					Text: "Discard a land card: Create a 2/2 green Bear creature token.",
					AdditionalCosts: []cost.Additional{
						cost.Additional{
							Kind:          cost.AdditionalDiscard,
							Text:          "Discard a land card",
							Amount:        1,
							MatchCardType: true,
							CardType:      types.Land,
							Source:        zone.Hand,
							CounterKind:   counter.PlusOnePlusOne,
						},
					},
					ZoneOfFunction: zone.Battlefield,
					Content: game.AbilityContent{
						Modes: []game.Mode{
							game.Mode{
								Sequence: []game.Instruction{
									game.Instruction{
										Primitive: game.CreateToken{
											Amount: game.Fixed(1),
											Source: game.TokenDef(ayulaSInfluenceToken),
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
			Discard a land card: Create a 2/2 green Bear creature token.
		`,
		},
	}
}

var ayulaSInfluenceToken = newAyulaSInfluenceToken()

func newAyulaSInfluenceToken() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name:      "Bear",
			Colors:    []color.Color{color.Green},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Bear},
			Power:     opt.Val(game.PT{Value: 2}),
			Toughness: opt.Val(game.PT{Value: 2}),
		},
	}
}
