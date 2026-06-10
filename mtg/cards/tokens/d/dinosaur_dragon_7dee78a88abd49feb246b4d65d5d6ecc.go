package d

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// Dinosaur Dragon
//
// Type: Token Creature — Dinosaur Dragon
//
// Oracle text:
//   Flying

// DinosaurDragonToken7dee78a88abd49feb246b4d65d5d6ecc is the card definition for Dinosaur Dragon.
var DinosaurDragonToken7dee78a88abd49feb246b4d65d5d6ecc = &game.CardDef{
	ColorIdentity: color.NewIdentity(color.Red),
	CardFace: game.CardFace{
		Name:      "Dinosaur Dragon",
		Colors:    []color.Color{color.Red},
		Types:     []types.Card{types.Creature},
		Subtypes:  []types.Sub{types.Dinosaur, types.Dragon},
		Power:     opt.Val(game.PT{Value: 4}),
		Toughness: opt.Val(game.PT{Value: 4}),
		StaticAbilities: []game.StaticAbility{
			game.FlyingStaticBody,
		},
		OracleText: `
			Flying
		`,
	},
}
