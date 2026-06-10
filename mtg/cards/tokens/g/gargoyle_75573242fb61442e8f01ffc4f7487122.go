package g

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// Gargoyle
//
// Type: Token Artifact Creature — Gargoyle
//
// Oracle text:
//   Flying

// GargoyleToken75573242fb61442e8f01ffc4f7487122 is the card definition for Gargoyle.
var GargoyleToken75573242fb61442e8f01ffc4f7487122 = &game.CardDef{
	CardFace: game.CardFace{
		Name:      "Gargoyle",
		Types:     []types.Card{types.Artifact, types.Creature},
		Subtypes:  []types.Sub{types.Gargoyle},
		Power:     opt.Val(game.PT{Value: 3}),
		Toughness: opt.Val(game.PT{Value: 4}),
		StaticAbilities: []game.StaticAbility{
			game.FlyingStaticBody,
		},
		OracleText: `
			Flying
		`,
	},
}
