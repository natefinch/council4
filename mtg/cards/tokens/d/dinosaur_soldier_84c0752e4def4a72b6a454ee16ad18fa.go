package d

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// Dinosaur Soldier
//
// Type: Token Creature — Dinosaur Soldier
//
// Oracle text:

// DinosaurSoldierToken84c0752e4def4a72b6a454ee16ad18fa is the card definition for Dinosaur Soldier.
var DinosaurSoldierToken84c0752e4def4a72b6a454ee16ad18fa = &game.CardDef{
	ColorIdentity: color.NewIdentity(color.White),
	CardFace: game.CardFace{
		Name:      "Dinosaur Soldier",
		Colors:    []color.Color{color.White},
		Types:     []types.Card{types.Creature},
		Subtypes:  []types.Sub{types.Dinosaur, types.Soldier},
		Power:     opt.Val(game.PT{Value: 2}),
		Toughness: opt.Val(game.PT{Value: 2}),
	},
}
