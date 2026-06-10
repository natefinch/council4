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

// SpiritToken399bc55311f7413c83c54b022fc6b3de is the card definition for Spirit.
var SpiritToken399bc55311f7413c83c54b022fc6b3de = &game.CardDef{
	CardFace: game.CardFace{
		Name:      "Spirit",
		Types:     []types.Card{types.Creature},
		Subtypes:  []types.Sub{types.Spirit},
		Power:     opt.Val(game.PT{Value: 2}),
		Toughness: opt.Val(game.PT{Value: 2}),
	},
}
