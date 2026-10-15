package r

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// RhoxFaithmender is the card definition for Rhox Faithmender.
//
// Type: Creature — Rhino Monk
// Cost: {3}{W}
//
// Oracle text:
//
//	Lifelink (Damage dealt by this creature also causes you to gain that much life.)
//	If you would gain life, you gain twice that much life instead.
var RhoxFaithmender = newRhoxFaithmender

func newRhoxFaithmender() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White),
		CardFace: game.CardFace{
			Name: "Rhox Faithmender",
			ManaCost: opt.Val(cost.Mana{
				cost.O(3),
				cost.W,
			}),
			Colors:    []color.Color{color.White},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Rhino, types.Monk},
			Power:     opt.Val(game.PT{Value: 1}),
			Toughness: opt.Val(game.PT{Value: 5}),
			StaticAbilities: []game.StaticAbility{
				game.LifelinkStaticBody,
			},
			ReplacementAbilities: []game.ReplacementAbility{
				game.LifeGainReplacement("If you would gain life, you gain twice that much life instead.", 2, 0),
			},
			OracleText: `
			Lifelink (Damage dealt by this creature also causes you to gain that much life.)
			If you would gain life, you gain twice that much life instead.
		`,
		},
	}
}
