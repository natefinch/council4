package s

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// StoneBoulder is the card definition for Stone Boulder.
//
// Type: Token Artifact Creature — Wall
//
// Oracle text:
//   Defender

// StoneBoulderToken48871e9ae08b4d4db8ab635ea7299e6e is the card definition for Stone Boulder.
var StoneBoulderToken48871e9ae08b4d4db8ab635ea7299e6e = newStoneBoulderToken48871e9ae08b4d4db8ab635ea7299e6e()

func newStoneBoulderToken48871e9ae08b4d4db8ab635ea7299e6e() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name:      "Stone Boulder",
			Types:     []types.Card{types.Artifact, types.Creature},
			Subtypes:  []types.Sub{types.Wall},
			Power:     opt.Val(game.PT{Value: 3}),
			Toughness: opt.Val(game.PT{Value: 1}),
			StaticAbilities: []game.StaticAbility{
				game.DefenderStaticBody,
			},
			OracleText: `
			Defender
		`,
		},
	}
}
