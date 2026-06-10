package z

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// Zombie Druid
//
// Type: Token Creature — Zombie Druid
//
// Oracle text:

// ZombieDruidTokend2c8634ae0b54c9eb720be8b18fc49dc is the card definition for Zombie Druid.
var ZombieDruidTokend2c8634ae0b54c9eb720be8b18fc49dc = &game.CardDef{
	ColorIdentity: color.NewIdentity(color.Black),
	CardFace: game.CardFace{
		Name:      "Zombie Druid",
		Colors:    []color.Color{color.Black},
		Types:     []types.Card{types.Creature},
		Subtypes:  []types.Sub{types.Zombie, types.Druid},
		Power:     opt.Val(game.PT{Value: 2}),
		Toughness: opt.Val(game.PT{Value: 2}),
	},
}
