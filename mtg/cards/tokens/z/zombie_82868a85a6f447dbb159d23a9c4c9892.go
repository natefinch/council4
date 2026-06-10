package z

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// Zombie
//
// Type: Token Artifact Creature — Zombie
//
// Oracle text:

// ZombieToken82868a85a6f447dbb159d23a9c4c9892 is the card definition for Zombie.
var ZombieToken82868a85a6f447dbb159d23a9c4c9892 = &game.CardDef{
	CardFace: game.CardFace{
		Name:      "Zombie",
		Types:     []types.Card{types.Artifact, types.Creature},
		Subtypes:  []types.Sub{types.Zombie},
		Power:     opt.Val(game.PT{Value: 3}),
		Toughness: opt.Val(game.PT{Value: 3}),
	},
}
