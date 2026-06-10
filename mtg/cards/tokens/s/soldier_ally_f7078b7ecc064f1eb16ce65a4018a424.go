package s

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// Soldier Ally
//
// Type: Token Creature — Soldier Ally
//
// Oracle text:

// SoldierAllyTokenf7078b7ecc064f1eb16ce65a4018a424 is the card definition for Soldier Ally.
var SoldierAllyTokenf7078b7ecc064f1eb16ce65a4018a424 = &game.CardDef{
	ColorIdentity: color.NewIdentity(color.White),
	CardFace: game.CardFace{
		Name:      "Soldier Ally",
		Colors:    []color.Color{color.White},
		Types:     []types.Card{types.Creature},
		Subtypes:  []types.Sub{types.Soldier, types.Ally},
		Power:     opt.Val(game.PT{Value: 1}),
		Toughness: opt.Val(game.PT{Value: 1}),
	},
}
