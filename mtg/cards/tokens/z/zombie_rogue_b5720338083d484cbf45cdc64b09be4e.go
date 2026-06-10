package z

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// Zombie Rogue
//
// Type: Token Creature — Zombie Rogue
//
// Oracle text:

// ZombieRogueTokenb5720338083d484cbf45cdc64b09be4e is the card definition for Zombie Rogue.
var ZombieRogueTokenb5720338083d484cbf45cdc64b09be4e = &game.CardDef{
	ColorIdentity: color.NewIdentity(color.Blue, color.Black),
	CardFace: game.CardFace{
		Name:      "Zombie Rogue",
		Colors:    []color.Color{color.Black, color.Blue},
		Types:     []types.Card{types.Creature},
		Subtypes:  []types.Sub{types.Zombie, types.Rogue},
		Power:     opt.Val(game.PT{Value: 2}),
		Toughness: opt.Val(game.PT{Value: 2}),
	},
}
