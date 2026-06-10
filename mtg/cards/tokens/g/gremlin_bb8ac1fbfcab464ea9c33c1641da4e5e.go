package g

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// Gremlin
//
// Type: Token Creature — Gremlin
//
// Oracle text:

// GremlinTokenbb8ac1fbfcab464ea9c33c1641da4e5e is the card definition for Gremlin.
var GremlinTokenbb8ac1fbfcab464ea9c33c1641da4e5e = &game.CardDef{
	ColorIdentity: color.NewIdentity(color.Red),
	CardFace: game.CardFace{
		Name:      "Gremlin",
		Colors:    []color.Color{color.Red},
		Types:     []types.Card{types.Creature},
		Subtypes:  []types.Sub{types.Gremlin},
		Power:     opt.Val(game.PT{Value: 2}),
		Toughness: opt.Val(game.PT{Value: 2}),
	},
}
