package k

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// Kor Ally
//
// Type: Token Creature — Kor Ally
//
// Oracle text:

// KorAllyTokendf5630cc9b854761af904f5c9d902127 is the card definition for Kor Ally.
var KorAllyTokendf5630cc9b854761af904f5c9d902127 = &game.CardDef{
	ColorIdentity: color.NewIdentity(color.White),
	CardFace: game.CardFace{
		Name:      "Kor Ally",
		Colors:    []color.Color{color.White},
		Types:     []types.Card{types.Creature},
		Subtypes:  []types.Sub{types.Kor, types.Ally},
		Power:     opt.Val(game.PT{Value: 1}),
		Toughness: opt.Val(game.PT{Value: 1}),
	},
}
