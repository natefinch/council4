package z

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// Zombie Berserker
//
// Type: Token Creature — Zombie Berserker
//
// Oracle text:

// ZombieBerserkerTokenb963103799e94cb791c2fe7b23bb5d4b is the card definition for Zombie Berserker.
var ZombieBerserkerTokenb963103799e94cb791c2fe7b23bb5d4b = &game.CardDef{
	ColorIdentity: color.NewIdentity(color.Black),
	CardFace: game.CardFace{
		Name:      "Zombie Berserker",
		Colors:    []color.Color{color.Black},
		Types:     []types.Card{types.Creature},
		Subtypes:  []types.Sub{types.Zombie, types.Berserker},
		Power:     opt.Val(game.PT{Value: 2}),
		Toughness: opt.Val(game.PT{Value: 2}),
	},
}
