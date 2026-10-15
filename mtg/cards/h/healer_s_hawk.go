package h

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// HealerSHawk is the card definition for Healer's Hawk.
//
// Type: Creature — Bird
// Cost: {W}
//
// Oracle text:
//
//	Flying
//	Lifelink (Damage dealt by this creature also causes you to gain that much life.)
var HealerSHawk = newHealerSHawk

func newHealerSHawk() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White),
		CardFace: game.CardFace{
			Name: "Healer's Hawk",
			ManaCost: opt.Val(cost.Mana{
				cost.W,
			}),
			Colors:    []color.Color{color.White},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Bird},
			Power:     opt.Val(game.PT{Value: 1}),
			Toughness: opt.Val(game.PT{Value: 1}),
			StaticAbilities: []game.StaticAbility{
				game.FlyingStaticBody,
				game.LifelinkStaticBody,
			},
			OracleText: `
			Flying
			Lifelink (Damage dealt by this creature also causes you to gain that much life.)
		`,
		},
	}
}
