package m

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// Mechtitan is the card definition for Mechtitan.
//
// Type: Token Legendary Artifact Creature — Construct
//
// Oracle text:
//   Mechtitan is all colors.
//   Flying, vigilance, trample, lifelink, haste

// MechtitanTokena4fecf0aa7b349e4bbb29690dff5a8f4 is the card definition for Mechtitan.
var MechtitanTokena4fecf0aa7b349e4bbb29690dff5a8f4 = newMechtitanTokena4fecf0aa7b349e4bbb29690dff5a8f4()

func newMechtitanTokena4fecf0aa7b349e4bbb29690dff5a8f4() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White, color.Blue, color.Black, color.Red, color.Green),
		CardFace: game.CardFace{
			Name:       "Mechtitan",
			Colors:     []color.Color{color.Black, color.Green, color.Red, color.Blue, color.White},
			Supertypes: []types.Super{types.Legendary},
			Types:      []types.Card{types.Artifact, types.Creature},
			Subtypes:   []types.Sub{types.Construct},
			Power:      opt.Val(game.PT{Value: 10}),
			Toughness:  opt.Val(game.PT{Value: 10}),
			StaticAbilities: []game.StaticAbility{
				game.StaticAbility{
					Text: "Mechtitan is all colors.",
					ContinuousEffects: []game.ContinuousEffect{
						game.ContinuousEffect{
							Duration:       game.DurationPermanent,
							AffectedSource: true,
							Layer:          game.LayerColor,
							SetColors:      []color.Color{color.White, color.Blue, color.Black, color.Red, color.Green},
						},
					},
				},
				game.FlyingStaticBody,
				game.VigilanceStaticBody,
				game.TrampleStaticBody,
				game.LifelinkStaticBody,
				game.HasteStaticBody,
			},
			OracleText: `
			Mechtitan is all colors.
			Flying, vigilance, trample, lifelink, haste
		`,
		},
	}
}
