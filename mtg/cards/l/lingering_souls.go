package l

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// LingeringSouls is the card definition for Lingering Souls.
//
// Type: Sorcery
// Cost: {2}{W}
//
// Oracle text:
//
//	Create two 1/1 white Spirit creature tokens with flying.
//	Flashback {1}{B} (You may cast this card from your graveyard for its flashback cost. Then exile it.)
var LingeringSouls = newLingeringSouls

func newLingeringSouls() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White, color.Black),
		CardFace: game.CardFace{
			Name: "Lingering Souls",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
				cost.W,
			}),
			Colors: []color.Color{color.White},
			Types:  []types.Card{types.Sorcery},
			StaticAbilities: []game.StaticAbility{
				game.StaticAbility{
					Text: "Flashback {1}{B}",
					KeywordAbilities: []game.KeywordAbility{
						game.FlashbackKeyword{
							Cost: cost.Mana{
								cost.O(1),
								cost.B,
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
							Source: game.TokenDef(lingeringSoulsToken),
						},
					},
				},
			}.Ability()),
			OracleText: `
			Create two 1/1 white Spirit creature tokens with flying.
			Flashback {1}{B} (You may cast this card from your graveyard for its flashback cost. Then exile it.)
		`,
		},
	}
}

var lingeringSoulsToken = newLingeringSoulsToken()

func newLingeringSoulsToken() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name:      "Spirit",
			Colors:    []color.Color{color.White},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Spirit},
			Power:     opt.Val(game.PT{Value: 1}),
			Toughness: opt.Val(game.PT{Value: 1}),
			StaticAbilities: []game.StaticAbility{
				game.FlyingStaticBody,
			},
		},
	}
}
