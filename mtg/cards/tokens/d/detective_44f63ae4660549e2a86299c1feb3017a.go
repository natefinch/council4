package d

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// Detective
//
// Type: Token Creature — Detective
//
// Oracle text:

// DetectiveToken44f63ae4660549e2a86299c1feb3017a is the card definition for Detective.
var DetectiveToken44f63ae4660549e2a86299c1feb3017a = &game.CardDef{
	ColorIdentity: color.NewIdentity(color.White, color.Blue),
	CardFace: game.CardFace{
		Name:      "Detective",
		Colors:    []color.Color{color.Blue, color.White},
		Types:     []types.Card{types.Creature},
		Subtypes:  []types.Sub{types.Detective},
		Power:     opt.Val(game.PT{Value: 2}),
		Toughness: opt.Val(game.PT{Value: 2}),
	},
}
