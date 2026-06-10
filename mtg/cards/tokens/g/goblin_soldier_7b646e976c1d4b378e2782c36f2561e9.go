package g

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// Goblin Soldier
//
// Type: Token Creature — Goblin Soldier
//
// Oracle text:

// GoblinSoldierToken7b646e976c1d4b378e2782c36f2561e9 is the card definition for Goblin Soldier.
var GoblinSoldierToken7b646e976c1d4b378e2782c36f2561e9 = &game.CardDef{
	ColorIdentity: color.NewIdentity(color.White, color.Red),
	CardFace: game.CardFace{
		Name:      "Goblin Soldier",
		Colors:    []color.Color{color.Red, color.White},
		Types:     []types.Card{types.Creature},
		Subtypes:  []types.Sub{types.Goblin, types.Soldier},
		Power:     opt.Val(game.PT{Value: 1}),
		Toughness: opt.Val(game.PT{Value: 1}),
	},
}
