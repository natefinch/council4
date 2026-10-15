package r

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// ReapTheSeagraf is the card definition for Reap the Seagraf.
//
// Type: Sorcery
// Cost: {2}{B}
//
// Oracle text:
//
//	Create a 2/2 black Zombie creature token.
//	Flashback {4}{U} (You may cast this card from your graveyard for its flashback cost. Then exile it.)
var ReapTheSeagraf = newReapTheSeagraf

func newReapTheSeagraf() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue, color.Black),
		CardFace: game.CardFace{
			Name: "Reap the Seagraf",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
				cost.B,
			}),
			Colors: []color.Color{color.Black},
			Types:  []types.Card{types.Sorcery},
			StaticAbilities: []game.StaticAbility{
				game.StaticAbility{
					Text: "Flashback {4}{U}",
					KeywordAbilities: []game.KeywordAbility{
						game.FlashbackKeyword{
							Cost: cost.Mana{
								cost.O(4),
								cost.U,
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
							Source: game.TokenDef(reapTheSeagrafToken),
						},
					},
				},
			}.Ability()),
			OracleText: `
			Create a 2/2 black Zombie creature token.
			Flashback {4}{U} (You may cast this card from your graveyard for its flashback cost. Then exile it.)
		`,
		},
	}
}

var reapTheSeagrafToken = newReapTheSeagrafToken()

func newReapTheSeagrafToken() *game.CardDef {
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
