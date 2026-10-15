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
//   Flying

// AngelToken350305763f03417aa52ac4356b78362f is the card definition for Angel.
var AngelToken350305763f03417aa52ac4356b78362f = newAngelToken350305763f03417aa52ac4356b78362f()

func newAngelToken350305763f03417aa52ac4356b78362f() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue),
		CardFace: game.CardFace{
			Name:      "Angel",
			Colors:    []color.Color{color.Blue},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Angel},
			Power:     opt.Val(game.PT{Value: 3}),
			Toughness: opt.Val(game.PT{Value: 3}),
			StaticAbilities: []game.StaticAbility{
				game.FlyingStaticBody,
			},
			OracleText: `
			Flying
		`,
		},
	}
}
