package p

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// PhyrexianMite is the card definition for Phyrexian Mite.
//
// Type: Token Artifact Creature — Phyrexian Mite
//
// Oracle text:
//   Toxic 1 (Players dealt combat damage by this creature also get a poison counter.)
//   This creature can't block.

// PhyrexianMiteToken2667d72301c84ea3ac17cedb3b842c3b is the card definition for Phyrexian Mite.
var PhyrexianMiteToken2667d72301c84ea3ac17cedb3b842c3b = newPhyrexianMiteToken2667d72301c84ea3ac17cedb3b842c3b()

func newPhyrexianMiteToken2667d72301c84ea3ac17cedb3b842c3b() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name:      "Phyrexian Mite",
			Types:     []types.Card{types.Artifact, types.Creature},
			Subtypes:  []types.Sub{types.Phyrexian, types.Mite},
			Power:     opt.Val(game.PT{Value: 1}),
			Toughness: opt.Val(game.PT{Value: 1}),
			StaticAbilities: []game.StaticAbility{
				game.StaticAbility{
					Text: "Toxic 1",
					KeywordAbilities: []game.KeywordAbility{
						game.ToxicKeyword{
							Amount: 1,
						},
					},
				},
				game.CantBlockStaticBody,
			},
			OracleText: `
			Toxic 1 (Players dealt combat damage by this creature also get a poison counter.)
			This creature can't block.
		`,
		},
	}
}
