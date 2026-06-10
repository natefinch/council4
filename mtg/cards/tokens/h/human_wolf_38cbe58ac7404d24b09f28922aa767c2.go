package h

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// Human // Wolf
//
// Type: Token Creature — Human // Token Creature — Wolf
// Face: Wolf — Token Creature — Wolf
//
// Oracle text:
//   Wolf
//

// HumanToken38cbe58ac7404d24b09f28922aa767c2 is the card definition for Human.
var HumanToken38cbe58ac7404d24b09f28922aa767c2 = &game.CardDef{
	ColorIdentity: color.NewIdentity(color.White, color.Green),
	CardFace: game.CardFace{
		Name:      "Human",
		Colors:    []color.Color{color.White},
		Types:     []types.Card{types.Creature},
		Subtypes:  []types.Sub{types.Human},
		Power:     opt.Val(game.PT{Value: 1}),
		Toughness: opt.Val(game.PT{Value: 1}),
	},
	Layout: game.LayoutDoubleFacedToken,
	Back: opt.Val(game.CardFace{
		Name:      "Wolf",
		Colors:    []color.Color{color.Green},
		Types:     []types.Card{types.Creature},
		Subtypes:  []types.Sub{types.Wolf},
		Power:     opt.Val(game.PT{Value: 2}),
		Toughness: opt.Val(game.PT{Value: 2}),
	}),
}
