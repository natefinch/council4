package g

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// Giant
//
// Type: Token Creature — Giant
//
// Oracle text:

// GiantToken33ec38946ec24038bed485e5b9cfe182 is the card definition for Giant.
var GiantToken33ec38946ec24038bed485e5b9cfe182 = &game.CardDef{
	ColorIdentity: color.NewIdentity(color.Green),
	CardFace: game.CardFace{
		Name:      "Giant",
		Colors:    []color.Color{color.Green},
		Types:     []types.Card{types.Creature},
		Subtypes:  []types.Sub{types.Giant},
		Power:     opt.Val(game.PT{Value: 7}),
		Toughness: opt.Val(game.PT{Value: 7}),
	},
}
