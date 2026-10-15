package s

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// SoltariFootSoldier is the card definition for Soltari Foot Soldier.
//
// Type: Creature — Soltari Soldier
// Cost: {W}
//
// Oracle text:
//
//	Shadow (This creature can block or be blocked by only creatures with shadow.)
var SoltariFootSoldier = newSoltariFootSoldier

func newSoltariFootSoldier() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White),
		CardFace: game.CardFace{
			Name: "Soltari Foot Soldier",
			ManaCost: opt.Val(cost.Mana{
				cost.W,
			}),
			Colors:    []color.Color{color.White},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Soltari, types.Soldier},
			Power:     opt.Val(game.PT{Value: 1}),
			Toughness: opt.Val(game.PT{Value: 1}),
			StaticAbilities: []game.StaticAbility{
				game.ShadowStaticBody,
			},
			OracleText: `
			Shadow (This creature can block or be blocked by only creatures with shadow.)
		`,
		},
	}
}
