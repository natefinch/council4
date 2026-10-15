package a

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// Angel is the card definition for Angel.
//
// Type: Token Creature — Angel
//
// Oracle text:
//   Flying, vigilance

// AngelToken9d256142cae245d2ab8f5b390b403481 is the card definition for Angel.
var AngelToken9d256142cae245d2ab8f5b390b403481 = newAngelToken9d256142cae245d2ab8f5b390b403481()

func newAngelToken9d256142cae245d2ab8f5b390b403481() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name:      "Angel",
			Colors:    []color.Color{color.White},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Angel},
			Power:     opt.Val(game.PT{Value: 3}),
			Toughness: opt.Val(game.PT{Value: 3}),
			StaticAbilities: []game.StaticAbility{
				game.FlyingStaticBody,
				game.VigilanceStaticBody,
			},
			OracleText: `
			Flying, vigilance
		`,
		},
	}
}
