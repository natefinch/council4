package a

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// AvenWindGuide is the card definition for Aven Wind Guide.
//
// Type: Token Creature — Zombie Bird Warrior
//
// Oracle text:
//   Flying, vigilance
//   Creature tokens you control have flying and vigilance.

// AvenWindGuideToken57a8745a5dd345f59f223b68d8d60618 is the card definition for Aven Wind Guide.
var AvenWindGuideToken57a8745a5dd345f59f223b68d8d60618 = newAvenWindGuideToken57a8745a5dd345f59f223b68d8d60618()

func newAvenWindGuideToken57a8745a5dd345f59f223b68d8d60618() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White),
		CardFace: game.CardFace{
			Name:      "Aven Wind Guide",
			Colors:    []color.Color{color.White},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Zombie, types.Bird, types.Warrior},
			Power:     opt.Val(game.PT{Value: 2}),
			Toughness: opt.Val(game.PT{Value: 3}),
			StaticAbilities: []game.StaticAbility{
				game.FlyingStaticBody,
				game.VigilanceStaticBody,
				game.StaticAbility{
					Text: "Creature tokens you control have flying and vigilance.",
					ContinuousEffects: []game.ContinuousEffect{
						game.ContinuousEffect{
							Duration:    game.DurationPermanent,
							Group:       game.ObjectControlledGroup(game.SourcePermanentReference(), game.Selection{RequiredTypes: []types.Card{types.Creature}, TokenOnly: true}),
							Layer:       game.LayerAbility,
							AddKeywords: []game.Keyword{game.Flying, game.Vigilance},
						},
					},
				},
			},
			OracleText: `
			Flying, vigilance
			Creature tokens you control have flying and vigilance.
		`,
		},
	}
}
