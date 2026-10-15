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
//   Changeling

// ShapeshifterTokenfe8478ec2af9472bb104a71521a3559f is the card definition for Shapeshifter.
var ShapeshifterTokenfe8478ec2af9472bb104a71521a3559f = newShapeshifterTokenfe8478ec2af9472bb104a71521a3559f()

func newShapeshifterTokenfe8478ec2af9472bb104a71521a3559f() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name:      "Shapeshifter",
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Shapeshifter},
			Power:     opt.Val(game.PT{Value: 1}),
			Toughness: opt.Val(game.PT{Value: 1}),
			StaticAbilities: []game.StaticAbility{
				game.ChangelingStaticBody,
			},
			OracleText: `
			Changeling
		`,
		},
	}
}
