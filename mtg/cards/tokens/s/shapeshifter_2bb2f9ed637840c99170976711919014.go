package s

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// Shapeshifter is the card definition for Shapeshifter.
//
// Type: Token Creature — Shapeshifter
//
// Oracle text:
//   Changeling (This token is every creature type.)

// ShapeshifterToken2bb2f9ed637840c99170976711919014 is the card definition for Shapeshifter.
var ShapeshifterToken2bb2f9ed637840c99170976711919014 = newShapeshifterToken2bb2f9ed637840c99170976711919014()

func newShapeshifterToken2bb2f9ed637840c99170976711919014() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue),
		CardFace: game.CardFace{
			Name:      "Shapeshifter",
			Colors:    []color.Color{color.Blue},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Shapeshifter},
			Power:     opt.Val(game.PT{Value: 2}),
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
