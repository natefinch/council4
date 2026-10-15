package s

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// SporeBurst is the card definition for Spore Burst.
//
// Type: Sorcery
// Cost: {3}{G}
//
// Oracle text:
//
//	Domain — Create a 1/1 green Saproling creature token for each basic land type among lands you control.
var SporeBurst = newSporeBurst

func newSporeBurst() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Green),
		CardFace: game.CardFace{
			Name: "Spore Burst",
			ManaCost: opt.Val(cost.Mana{
				cost.O(3),
				cost.G,
			}),
			Colors: []color.Color{color.Green},
			Types:  []types.Card{types.Sorcery},
			SpellAbility: opt.Val(game.Mode{
				Sequence: []game.Instruction{
					{
						Primitive: game.CreateToken{
							Amount: game.Dynamic(game.DynamicAmount{
								Kind:       game.DynamicAmountControllerBasicLandTypeCount,
								Multiplier: 1,
							}),
							Source: game.TokenDef(sporeBurstToken),
						},
					},
				},
			}.Ability()),
			OracleText: `
			Domain — Create a 1/1 green Saproling creature token for each basic land type among lands you control.
		`,
		},
	}
}

var sporeBurstToken = newSporeBurstToken()

func newSporeBurstToken() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name:      "Saproling",
			Colors:    []color.Color{color.Green},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Saproling},
			Power:     opt.Val(game.PT{Value: 1}),
			Toughness: opt.Val(game.PT{Value: 1}),
		},
	}
}
