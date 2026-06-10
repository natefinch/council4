package o

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// Orc Army
//
// Type: Token Creature — Orc Army
//
// Oracle text:

// OrcArmyToken620159f8eb7f4756a1368ad9728adbad is the card definition for Orc Army.
var OrcArmyToken620159f8eb7f4756a1368ad9728adbad = &game.CardDef{
	ColorIdentity: color.NewIdentity(color.Black),
	CardFace: game.CardFace{
		Name:      "Orc Army",
		Colors:    []color.Color{color.Black},
		Types:     []types.Card{types.Creature},
		Subtypes:  []types.Sub{types.Orc, types.Army},
		Power:     opt.Val(game.PT{Value: 0}),
		Toughness: opt.Val(game.PT{Value: 0}),
	},
}
