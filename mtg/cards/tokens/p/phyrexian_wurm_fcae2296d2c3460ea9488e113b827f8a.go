package p

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// Phyrexian Wurm
//
// Type: Token Artifact Creature — Phyrexian Wurm
//
// Oracle text:
//   Deathtouch

// PhyrexianWurmTokenfcae2296d2c3460ea9488e113b827f8a is the card definition for Phyrexian Wurm.
var PhyrexianWurmTokenfcae2296d2c3460ea9488e113b827f8a = &game.CardDef{
	ColorIdentity: color.NewIdentity(color.Black),
	CardFace: game.CardFace{
		Name:      "Phyrexian Wurm",
		Colors:    []color.Color{color.Black},
		Types:     []types.Card{types.Artifact, types.Creature},
		Subtypes:  []types.Sub{types.Phyrexian, types.Wurm},
		Power:     opt.Val(game.PT{Value: 1}),
		Toughness: opt.Val(game.PT{Value: 2}),
		StaticAbilities: []game.StaticAbility{
			game.DeathtouchStaticBody,
		},
		OracleText: `
			Deathtouch
		`,
	},
}
