package s

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// Soldier // Goblin
//
// Type: Token Creature — Soldier // Token Creature — Goblin
// Face: Goblin — Token Creature — Goblin
//
// Oracle text:
//   Goblin
//

// SoldierTokenb973cb863a1b42de83ae74dffa5bcb8a is the card definition for Soldier.
var SoldierTokenb973cb863a1b42de83ae74dffa5bcb8a = &game.CardDef{
	ColorIdentity: color.NewIdentity(color.White, color.Red),
	CardFace: game.CardFace{
		Name:      "Soldier",
		Colors:    []color.Color{color.Red, color.White},
		Types:     []types.Card{types.Creature},
		Subtypes:  []types.Sub{types.Soldier},
		Power:     opt.Val(game.PT{Value: 1}),
		Toughness: opt.Val(game.PT{Value: 1}),
	},
	Layout: game.LayoutDoubleFacedToken,
	Back: opt.Val(game.CardFace{
		Name:      "Goblin",
		Colors:    []color.Color{color.Red},
		Types:     []types.Card{types.Creature},
		Subtypes:  []types.Sub{types.Goblin},
		Power:     opt.Val(game.PT{Value: 1}),
		Toughness: opt.Val(game.PT{Value: 1}),
	}),
}
