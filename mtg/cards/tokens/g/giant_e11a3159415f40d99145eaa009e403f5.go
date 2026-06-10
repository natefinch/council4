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

// GiantTokene11a3159415f40d99145eaa009e403f5 is the card definition for Giant.
var GiantTokene11a3159415f40d99145eaa009e403f5 = &game.CardDef{
	ColorIdentity: color.NewIdentity(color.Red),
	CardFace: game.CardFace{
		Name:      "Giant",
		Colors:    []color.Color{color.Red},
		Types:     []types.Card{types.Creature},
		Subtypes:  []types.Sub{types.Giant},
		Power:     opt.Val(game.PT{Value: 4}),
		Toughness: opt.Val(game.PT{Value: 4}),
	},
}
