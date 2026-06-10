package k

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// Knight Ally
//
// Type: Token Creature — Knight Ally
//
// Oracle text:

// KnightAllyToken289f88eb87a24927b6d10ff1fc4b2b2f is the card definition for Knight Ally.
var KnightAllyToken289f88eb87a24927b6d10ff1fc4b2b2f = &game.CardDef{
	ColorIdentity: color.NewIdentity(color.White),
	CardFace: game.CardFace{
		Name:      "Knight Ally",
		Colors:    []color.Color{color.White},
		Types:     []types.Card{types.Creature},
		Subtypes:  []types.Sub{types.Knight, types.Ally},
		Power:     opt.Val(game.PT{Value: 2}),
		Toughness: opt.Val(game.PT{Value: 2}),
	},
}
