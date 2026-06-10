package g

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// Goblin Rogue
//
// Type: Token Creature — Goblin Rogue
//
// Oracle text:

// GoblinRogueTokena0814e97bf284ffbb65d845de3613065 is the card definition for Goblin Rogue.
var GoblinRogueTokena0814e97bf284ffbb65d845de3613065 = &game.CardDef{
	ColorIdentity: color.NewIdentity(color.Black),
	CardFace: game.CardFace{
		Name:      "Goblin Rogue",
		Colors:    []color.Color{color.Black},
		Types:     []types.Card{types.Creature},
		Subtypes:  []types.Sub{types.Goblin, types.Rogue},
		Power:     opt.Val(game.PT{Value: 1}),
		Toughness: opt.Val(game.PT{Value: 1}),
	},
}
