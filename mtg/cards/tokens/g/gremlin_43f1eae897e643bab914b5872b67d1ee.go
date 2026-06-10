package g

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// Gremlin
//
// Type: Token Artifact Creature — Gremlin
//
// Oracle text:

// GremlinToken43f1eae897e643bab914b5872b67d1ee is the card definition for Gremlin.
var GremlinToken43f1eae897e643bab914b5872b67d1ee = &game.CardDef{
	ColorIdentity: color.NewIdentity(color.Red),
	CardFace: game.CardFace{
		Name:      "Gremlin",
		Colors:    []color.Color{color.Red},
		Types:     []types.Card{types.Artifact, types.Creature},
		Subtypes:  []types.Sub{types.Gremlin},
		Power:     opt.Val(game.PT{Value: 0}),
		Toughness: opt.Val(game.PT{Value: 0}),
	},
}
