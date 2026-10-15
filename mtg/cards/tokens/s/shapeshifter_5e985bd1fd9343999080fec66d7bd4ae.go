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

// ShapeshifterToken5e985bd1fd9343999080fec66d7bd4ae is the card definition for Shapeshifter.
var ShapeshifterToken5e985bd1fd9343999080fec66d7bd4ae = newShapeshifterToken5e985bd1fd9343999080fec66d7bd4ae()

func newShapeshifterToken5e985bd1fd9343999080fec66d7bd4ae() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name:      "Shapeshifter",
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Shapeshifter},
			Power:     opt.Val(game.PT{Value: 2}),
			Toughness: opt.Val(game.PT{Value: 2}),
			StaticAbilities: []game.StaticAbility{
				game.ChangelingStaticBody,
			},
			OracleText: `
			Changeling
		`,
		},
	}
}
