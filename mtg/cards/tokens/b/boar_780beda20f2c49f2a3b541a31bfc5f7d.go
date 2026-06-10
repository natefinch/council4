package b

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// Boar
//
// Type: Token Creature — Boar
//
// Oracle text:

// BoarToken780beda20f2c49f2a3b541a31bfc5f7d is the card definition for Boar.
var BoarToken780beda20f2c49f2a3b541a31bfc5f7d = &game.CardDef{
	ColorIdentity: color.NewIdentity(color.Green),
	CardFace: game.CardFace{
		Name:      "Boar",
		Colors:    []color.Color{color.Green},
		Types:     []types.Card{types.Creature},
		Subtypes:  []types.Sub{types.Boar},
		Power:     opt.Val(game.PT{Value: 3}),
		Toughness: opt.Val(game.PT{Value: 1}),
	},
}
