package k

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// Kavu is the card definition for Kavu.
//
// Type: Token Creature — Kavu
//
// Oracle text:
//   This creature is all colors.
//   Trample

// KavuToken9fef99020fbb4015922ead1ddc771da1 is the card definition for Kavu.
var KavuToken9fef99020fbb4015922ead1ddc771da1 = newKavuToken9fef99020fbb4015922ead1ddc771da1()

func newKavuToken9fef99020fbb4015922ead1ddc771da1() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White, color.Blue, color.Black, color.Red, color.Green),
		CardFace: game.CardFace{
			Name:      "Kavu",
			Colors:    []color.Color{color.Black, color.Green, color.Red, color.Blue, color.White},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Kavu},
			Power:     opt.Val(game.PT{Value: 3}),
			Toughness: opt.Val(game.PT{Value: 3}),
			StaticAbilities: []game.StaticAbility{
				game.StaticAbility{
					Text: "This creature is all colors.",
					ContinuousEffects: []game.ContinuousEffect{
						game.ContinuousEffect{
							Duration:       game.DurationPermanent,
							AffectedSource: true,
							Layer:          game.LayerColor,
							SetColors:      []color.Color{color.White, color.Blue, color.Black, color.Red, color.Green},
						},
					},
				},
				game.TrampleStaticBody,
			},
			OracleText: `
			This creature is all colors.
			Trample
		`,
		},
	}
}
