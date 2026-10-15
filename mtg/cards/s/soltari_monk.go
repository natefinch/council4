package s

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// SoltariMonk is the card definition for Soltari Monk.
//
// Type: Creature — Soltari Monk Cleric
// Cost: {W}{W}
//
// Oracle text:
//
//	Protection from black
//	Shadow (This creature can block or be blocked by only creatures with shadow.)
var SoltariMonk = newSoltariMonk

func newSoltariMonk() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White),
		CardFace: game.CardFace{
			Name: "Soltari Monk",
			ManaCost: opt.Val(cost.Mana{
				cost.W,
				cost.W,
			}),
			Colors:    []color.Color{color.White},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Soltari, types.Monk, types.Cleric},
			Power:     opt.Val(game.PT{Value: 2}),
			Toughness: opt.Val(game.PT{Value: 1}),
			StaticAbilities: []game.StaticAbility{
				game.ProtectionFromColorsStaticAbility(color.Black),
				game.ShadowStaticBody,
			},
			OracleText: `
			Protection from black
			Shadow (This creature can block or be blocked by only creatures with shadow.)
		`,
		},
	}
}
