package s

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// SphinxOfTheSteelWind is the card definition for Sphinx of the Steel Wind.
//
// Type: Artifact Creature — Sphinx
// Cost: {5}{W}{U}{B}
//
// Oracle text:
//
//	Flying, first strike, vigilance, lifelink, protection from red and from green
var SphinxOfTheSteelWind = newSphinxOfTheSteelWind

func newSphinxOfTheSteelWind() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White, color.Blue, color.Black),
		CardFace: game.CardFace{
			Name: "Sphinx of the Steel Wind",
			ManaCost: opt.Val(cost.Mana{
				cost.O(5),
				cost.W,
				cost.U,
				cost.B,
			}),
			Colors:    []color.Color{color.Black, color.Blue, color.White},
			Types:     []types.Card{types.Artifact, types.Creature},
			Subtypes:  []types.Sub{types.Sphinx},
			Power:     opt.Val(game.PT{Value: 6}),
			Toughness: opt.Val(game.PT{Value: 6}),
			StaticAbilities: []game.StaticAbility{
				game.FlyingStaticBody,
				game.FirstStrikeStaticBody,
				game.VigilanceStaticBody,
				game.LifelinkStaticBody,
				game.ProtectionFromColorsStaticAbility(color.Red, color.Green),
			},
			OracleText: `
			Flying, first strike, vigilance, lifelink, protection from red and from green
		`,
		},
	}
}
