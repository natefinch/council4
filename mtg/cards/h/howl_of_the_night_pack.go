package h

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// HowlOfTheNightPack is the card definition for Howl of the Night Pack.
//
// Type: Sorcery
// Cost: {6}{G}
//
// Oracle text:
//
//	Create a 2/2 green Wolf creature token for each Forest you control.
var HowlOfTheNightPack = newHowlOfTheNightPack

func newHowlOfTheNightPack() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Green),
		CardFace: game.CardFace{
			Name: "Howl of the Night Pack",
			ManaCost: opt.Val(cost.Mana{
				cost.O(6),
				cost.G,
			}),
			Colors: []color.Color{color.Green},
			Types:  []types.Card{types.Sorcery},
			SpellAbility: opt.Val(game.Mode{
				Sequence: []game.Instruction{
					{
						Primitive: game.CreateToken{
							Amount: game.Dynamic(game.DynamicAmount{
								Kind:       game.DynamicAmountCountSelector,
								Multiplier: 1,
								Group:      game.BattlefieldGroup(game.Selection{SubtypesAny: []types.Sub{types.Sub("Forest")}, Controller: game.ControllerYou}),
							}),
							Source: game.TokenDef(howlOfTheNightPackToken),
						},
					},
				},
			}.Ability()),
			OracleText: `
			Create a 2/2 green Wolf creature token for each Forest you control.
		`,
		},
	}
}

var howlOfTheNightPackToken = newHowlOfTheNightPackToken()

func newHowlOfTheNightPackToken() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name:      "Wolf",
			Colors:    []color.Color{color.Green},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Wolf},
			Power:     opt.Val(game.PT{Value: 2}),
			Toughness: opt.Val(game.PT{Value: 2}),
		},
	}
}
