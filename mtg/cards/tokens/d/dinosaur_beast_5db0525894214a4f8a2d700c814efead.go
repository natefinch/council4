package d

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// Dinosaur Beast
//
// Type: Token Creature — Dinosaur Beast
//
// Oracle text:
//   Trample

// DinosaurBeastToken5db0525894214a4f8a2d700c814efead is the card definition for Dinosaur Beast.
var DinosaurBeastToken5db0525894214a4f8a2d700c814efead = &game.CardDef{
	ColorIdentity: color.NewIdentity(color.Green),
	CardFace: game.CardFace{
		Name:      "Dinosaur Beast",
		Colors:    []color.Color{color.Green},
		Types:     []types.Card{types.Creature},
		Subtypes:  []types.Sub{types.Dinosaur, types.Beast},
		Power:     opt.Val(game.PT{IsStar: true}),
		Toughness: opt.Val(game.PT{IsStar: true}),
		StaticAbilities: []game.StaticAbility{
			game.TrampleStaticBody,
		},
		OracleText: `
			Trample
		`,
	},
}
