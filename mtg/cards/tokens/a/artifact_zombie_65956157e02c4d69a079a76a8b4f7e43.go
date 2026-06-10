package a

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// Artifact Zombie
//
// Type: Token Artifact Creature — Zombie
//
// Oracle text:

// ArtifactZombieToken65956157e02c4d69a079a76a8b4f7e43 is the card definition for Artifact Zombie.
var ArtifactZombieToken65956157e02c4d69a079a76a8b4f7e43 = &game.CardDef{
	CardFace: game.CardFace{
		Name:      "Artifact Zombie",
		Types:     []types.Card{types.Artifact, types.Creature},
		Subtypes:  []types.Sub{types.Zombie},
		Power:     opt.Val(game.PT{IsStar: true}),
		Toughness: opt.Val(game.PT{IsStar: true}),
	},
}
