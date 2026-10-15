package b

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// Beast is the card definition for Beast.
//
// Type: Token Creature — Beast
//
// Oracle text:
//   Trample

// BeastToken7ed8a016a2594fdaaae502f8df94f2de is the card definition for Beast.
var BeastToken7ed8a016a2594fdaaae502f8df94f2de = newBeastToken7ed8a016a2594fdaaae502f8df94f2de()

func newBeastToken7ed8a016a2594fdaaae502f8df94f2de() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Green),
		CardFace: game.CardFace{
			Name:      "Beast",
			Colors:    []color.Color{color.Green},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Beast},
			Power:     opt.Val(game.PT{Value: 4}),
			Toughness: opt.Val(game.PT{Value: 4}),
			StaticAbilities: []game.StaticAbility{
				game.TrampleStaticBody,
			},
			OracleText: `
			Trample
		`,
		},
	}
}
