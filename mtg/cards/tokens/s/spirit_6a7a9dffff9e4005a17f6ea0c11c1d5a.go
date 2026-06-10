package s

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// Spirit
//
// Type: Token Creature — Spirit
//
// Oracle text:

// SpiritToken6a7a9dffff9e4005a17f6ea0c11c1d5a is the card definition for Spirit.
var SpiritToken6a7a9dffff9e4005a17f6ea0c11c1d5a = &game.CardDef{
	CardFace: game.CardFace{
		Name:      "Spirit",
		Types:     []types.Card{types.Creature},
		Subtypes:  []types.Sub{types.Spirit},
		Power:     opt.Val(game.PT{Value: 1}),
		Toughness: opt.Val(game.PT{Value: 1}),
	},
}
