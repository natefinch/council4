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
//   Trample (This creature can deal excess combat damage to the player or planeswalker it's attacking.)

// DinosaurToken240300bd508843c5a873290507515843 is the card definition for Dinosaur.
var DinosaurToken240300bd508843c5a873290507515843 = &game.CardDef{
	ColorIdentity: color.NewIdentity(color.Green),
	CardFace: game.CardFace{
		Name:      "Dinosaur",
		Colors:    []color.Color{color.Green},
		Types:     []types.Card{types.Creature},
		Subtypes:  []types.Sub{types.Dinosaur},
		Power:     opt.Val(game.PT{Value: 3}),
		Toughness: opt.Val(game.PT{Value: 3}),
		StaticAbilities: []game.StaticAbility{
			game.TrampleStaticBody,
		},
		OracleText: `
			Trample (This creature can deal excess combat damage to the player or planeswalker it's attacking.)
		`,
	},
}
