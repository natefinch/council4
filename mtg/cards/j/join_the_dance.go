package j

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// JoinTheDance is the card definition for Join the Dance.
//
// Type: Sorcery
// Cost: {G}{W}
//
// Oracle text:
//
//	Create two 1/1 white Human creature tokens.
//	Flashback {3}{G}{W} (You may cast this card from your graveyard for its flashback cost. Then exile it.)
var JoinTheDance = newJoinTheDance

func newJoinTheDance() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White, color.Green),
		CardFace: game.CardFace{
			Name: "Join the Dance",
			ManaCost: opt.Val(cost.Mana{
				cost.G,
				cost.W,
			}),
			Colors: []color.Color{color.Green, color.White},
			Types:  []types.Card{types.Sorcery},
			StaticAbilities: []game.StaticAbility{
				game.StaticAbility{
					Text: "Flashback {3}{G}{W}",
					KeywordAbilities: []game.KeywordAbility{
						game.FlashbackKeyword{
							Cost: cost.Mana{
								cost.O(3),
								cost.G,
								cost.W,
							},
						},
					},
				},
			},
			SpellAbility: opt.Val(game.Mode{
				Sequence: []game.Instruction{
					{
						Primitive: game.CreateToken{
							Amount: game.Fixed(2),
							Source: game.TokenDef(joinTheDanceToken),
						},
					},
				},
			}.Ability()),
			OracleText: `
			Create two 1/1 white Human creature tokens.
			Flashback {3}{G}{W} (You may cast this card from your graveyard for its flashback cost. Then exile it.)
		`,
		},
	}
}

var joinTheDanceToken = newJoinTheDanceToken()

func newJoinTheDanceToken() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name:      "Human",
			Colors:    []color.Color{color.White},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Human},
			Power:     opt.Val(game.PT{Value: 1}),
			Toughness: opt.Val(game.PT{Value: 1}),
		},
	}
}
