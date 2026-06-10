package o

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// Otter
//
// Type: Token Creature — Otter
//
// Oracle text:
//   Prowess (Whenever you cast a noncreature spell, this creature gets +1/+1 until end of turn.)

// OtterToken6c1bb10eb0834978ab0a6f6a6b5bcf2c is the card definition for Otter.
var OtterToken6c1bb10eb0834978ab0a6f6a6b5bcf2c = &game.CardDef{
	ColorIdentity: color.NewIdentity(color.Blue, color.Red),
	CardFace: game.CardFace{
		Name:      "Otter",
		Colors:    []color.Color{color.Red, color.Blue},
		Types:     []types.Card{types.Creature},
		Subtypes:  []types.Sub{types.Otter},
		Power:     opt.Val(game.PT{Value: 1}),
		Toughness: opt.Val(game.PT{Value: 1}),
		StaticAbilities: []game.StaticAbility{
			game.ProwessStaticBody,
		},
		OracleText: `
			Prowess (Whenever you cast a noncreature spell, this creature gets +1/+1 until end of turn.)
		`,
	},
}
