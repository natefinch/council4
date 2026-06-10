package d

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// Demon Berserker
//
// Type: Token Creature — Demon Berserker
//
// Oracle text:
//   Menace (This creature can't be blocked except by two or more creatures.)

// DemonBerserkerToken93b52e8b2e8848f680269397878b1d72 is the card definition for Demon Berserker.
var DemonBerserkerToken93b52e8b2e8848f680269397878b1d72 = &game.CardDef{
	ColorIdentity: color.NewIdentity(color.Red),
	CardFace: game.CardFace{
		Name:      "Demon Berserker",
		Colors:    []color.Color{color.Red},
		Types:     []types.Card{types.Creature},
		Subtypes:  []types.Sub{types.Demon, types.Berserker},
		Power:     opt.Val(game.PT{Value: 2}),
		Toughness: opt.Val(game.PT{Value: 3}),
		StaticAbilities: []game.StaticAbility{
			game.MenaceStaticBody,
		},
		OracleText: `
			Menace (This creature can't be blocked except by two or more creatures.)
		`,
	},
}
