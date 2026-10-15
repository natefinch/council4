package b

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// BlurredMongoose is the card definition for Blurred Mongoose.
//
// Type: Creature — Mongoose
// Cost: {1}{G}
//
// Oracle text:
//
//	This spell can't be countered.
//	Shroud (This creature can't be the target of spells or abilities.)
var BlurredMongoose = newBlurredMongoose

func newBlurredMongoose() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Green),
		CardFace: game.CardFace{
			Name: "Blurred Mongoose",
			ManaCost: opt.Val(cost.Mana{
				cost.O(1),
				cost.G,
			}),
			Colors:    []color.Color{color.Green},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Mongoose},
			Power:     opt.Val(game.PT{Value: 2}),
			Toughness: opt.Val(game.PT{Value: 1}),
			StaticAbilities: []game.StaticAbility{
				game.CantBeCounteredStaticBody,
				game.ShroudStaticBody,
			},
			OracleText: `
			This spell can't be countered.
			Shroud (This creature can't be the target of spells or abilities.)
		`,
		},
	}
}
