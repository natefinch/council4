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

// ShapeshifterTokenb9ad70d07e714ec1a39e7d9363501322 is the card definition for Shapeshifter.
var ShapeshifterTokenb9ad70d07e714ec1a39e7d9363501322 = newShapeshifterTokenb9ad70d07e714ec1a39e7d9363501322()

func newShapeshifterTokenb9ad70d07e714ec1a39e7d9363501322() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name:      "Shapeshifter",
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Shapeshifter},
			Power:     opt.Val(game.PT{Value: 3}),
			Toughness: opt.Val(game.PT{Value: 2}),
			StaticAbilities: []game.StaticAbility{
				game.ChangelingStaticBody,
			},
			OracleText: `
			Changeling (This token is every creature type.)
		`,
		},
	}
}
