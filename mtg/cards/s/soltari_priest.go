package s

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// SoltariPriest is the card definition for Soltari Priest.
//
// Type: Creature — Soltari Cleric
// Cost: {W}{W}
//
// Oracle text:
//
//	Protection from red
//	Shadow (This creature can block or be blocked by only creatures with shadow.)
var SoltariPriest = newSoltariPriest

func newSoltariPriest() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White),
		CardFace: game.CardFace{
			Name: "Soltari Priest",
			ManaCost: opt.Val(cost.Mana{
				cost.W,
				cost.W,
			}),
			Colors:    []color.Color{color.White},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Soltari, types.Cleric},
			Power:     opt.Val(game.PT{Value: 2}),
			Toughness: opt.Val(game.PT{Value: 1}),
			StaticAbilities: []game.StaticAbility{
				game.ProtectionFromColorsStaticAbility(color.Red),
				game.ShadowStaticBody,
			},
			OracleText: `
			Protection from red
			Shadow (This creature can block or be blocked by only creatures with shadow.)
		`,
		},
	}
}
