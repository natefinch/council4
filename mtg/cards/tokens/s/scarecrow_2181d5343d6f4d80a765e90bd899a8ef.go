package s

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// Scarecrow
//
// Type: Token Artifact Creature — Scarecrow
//
// Oracle text:
//   Vigilance

// ScarecrowToken2181d5343d6f4d80a765e90bd899a8ef is the card definition for Scarecrow.
var ScarecrowToken2181d5343d6f4d80a765e90bd899a8ef = &game.CardDef{
	CardFace: game.CardFace{
		Name:      "Scarecrow",
		Types:     []types.Card{types.Artifact, types.Creature},
		Subtypes:  []types.Sub{types.Scarecrow},
		Power:     opt.Val(game.PT{Value: 4}),
		Toughness: opt.Val(game.PT{Value: 4}),
		StaticAbilities: []game.StaticAbility{
			game.VigilanceStaticBody,
		},
		OracleText: `
			Vigilance
		`,
	},
}
