package z

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// Zombie
//
// Type: Token Enchantment Creature — Zombie
//
// Oracle text:

// ZombieToken7cd9759978584fba8ccc06aed3246380 is the card definition for Zombie.
var ZombieToken7cd9759978584fba8ccc06aed3246380 = &game.CardDef{
	ColorIdentity: color.NewIdentity(color.Black),
	CardFace: game.CardFace{
		Name:      "Zombie",
		Colors:    []color.Color{color.Black},
		Types:     []types.Card{types.Enchantment, types.Creature},
		Subtypes:  []types.Sub{types.Zombie},
		Power:     opt.Val(game.PT{Value: 2}),
		Toughness: opt.Val(game.PT{Value: 2}),
	},
}
