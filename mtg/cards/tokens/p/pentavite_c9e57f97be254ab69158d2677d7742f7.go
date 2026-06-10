package p

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// Pentavite
//
// Type: Token Artifact Creature — Pentavite
//
// Oracle text:
//   Flying

// PentaviteTokenc9e57f97be254ab69158d2677d7742f7 is the card definition for Pentavite.
var PentaviteTokenc9e57f97be254ab69158d2677d7742f7 = &game.CardDef{
	CardFace: game.CardFace{
		Name:      "Pentavite",
		Types:     []types.Card{types.Artifact, types.Creature},
		Subtypes:  []types.Sub{types.Pentavite},
		Power:     opt.Val(game.PT{Value: 1}),
		Toughness: opt.Val(game.PT{Value: 1}),
		StaticAbilities: []game.StaticAbility{
			game.FlyingStaticBody,
		},
		OracleText: `
			Flying
		`,
	},
}
