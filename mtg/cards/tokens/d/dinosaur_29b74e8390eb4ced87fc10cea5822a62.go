package d

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// Dinosaur
//
// Type: Token Creature — Dinosaur
//
// Oracle text:
//   Haste

// DinosaurToken29b74e8390eb4ced87fc10cea5822a62 is the card definition for Dinosaur.
var DinosaurToken29b74e8390eb4ced87fc10cea5822a62 = &game.CardDef{
	ColorIdentity: color.NewIdentity(color.Red),
	CardFace: game.CardFace{
		Name:      "Dinosaur",
		Colors:    []color.Color{color.Red},
		Types:     []types.Card{types.Creature},
		Subtypes:  []types.Sub{types.Dinosaur},
		Power:     opt.Val(game.PT{Value: 1}),
		Toughness: opt.Val(game.PT{Value: 1}),
		StaticAbilities: []game.StaticAbility{
			game.HasteStaticBody,
		},
		OracleText: `
			Haste
		`,
	},
}
