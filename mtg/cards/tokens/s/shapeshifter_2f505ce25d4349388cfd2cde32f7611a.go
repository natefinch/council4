package s

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// Shapeshifter is the card definition for Shapeshifter.
//
// Type: Token Creature — Shapeshifter
//
// Oracle text:
//   Changeling (This token is every creature type.)
//   Deathtouch

// ShapeshifterToken2f505ce25d4349388cfd2cde32f7611a is the card definition for Shapeshifter.
var ShapeshifterToken2f505ce25d4349388cfd2cde32f7611a = newShapeshifterToken2f505ce25d4349388cfd2cde32f7611a()

func newShapeshifterToken2f505ce25d4349388cfd2cde32f7611a() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name:      "Shapeshifter",
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Shapeshifter},
			Power:     opt.Val(game.PT{IsStar: true}),
			Toughness: opt.Val(game.PT{IsStar: true}),
			StaticAbilities: []game.StaticAbility{
				game.ChangelingStaticBody,
				game.DeathtouchStaticBody,
			},
			OracleText: `
			Changeling (This token is every creature type.)
			Deathtouch
		`,
		},
	}
}
