package c

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// ChatterOfTheSquirrel is the card definition for Chatter of the Squirrel.
//
// Type: Sorcery
// Cost: {G}
//
// Oracle text:
//
//	Create a 1/1 green Squirrel creature token.
//	Flashback {1}{G} (You may cast this card from your graveyard for its flashback cost. Then exile it.)
var ChatterOfTheSquirrel = newChatterOfTheSquirrel

func newChatterOfTheSquirrel() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Green),
		CardFace: game.CardFace{
			Name: "Chatter of the Squirrel",
			ManaCost: opt.Val(cost.Mana{
				cost.G,
			}),
			Colors: []color.Color{color.Green},
			Types:  []types.Card{types.Sorcery},
			StaticAbilities: []game.StaticAbility{
				game.StaticAbility{
					Text: "Flashback {1}{G}",
					KeywordAbilities: []game.KeywordAbility{
						game.FlashbackKeyword{
							Cost: cost.Mana{
								cost.O(1),
								cost.G,
							},
						},
					},
				},
			},
			SpellAbility: opt.Val(game.Mode{
				Sequence: []game.Instruction{
					{
						Primitive: game.CreateToken{
							Amount: game.Fixed(1),
							Source: game.TokenDef(chatterOfTheSquirrelToken),
						},
					},
				},
			}.Ability()),
			OracleText: `
			Create a 1/1 green Squirrel creature token.
			Flashback {1}{G} (You may cast this card from your graveyard for its flashback cost. Then exile it.)
		`,
		},
	}
}

var chatterOfTheSquirrelToken = newChatterOfTheSquirrelToken()

func newChatterOfTheSquirrelToken() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name:      "Squirrel",
			Colors:    []color.Color{color.Green},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Squirrel},
			Power:     opt.Val(game.PT{Value: 1}),
			Toughness: opt.Val(game.PT{Value: 1}),
		},
	}
}
