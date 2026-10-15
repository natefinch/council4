package h

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// HealerSFlock is the card definition for Healer's Flock.
//
// Type: Creature — Bird
// Cost: {W}{W}{W}
//
// Oracle text:
//
//	Flying, lifelink
var HealerSFlock = newHealerSFlock

func newHealerSFlock() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White),
		CardFace: game.CardFace{
			Name: "Healer's Flock",
			ManaCost: opt.Val(cost.Mana{
				cost.W,
				cost.W,
				cost.W,
			}),
			Colors:    []color.Color{color.White},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Bird},
			Power:     opt.Val(game.PT{Value: 3}),
			Toughness: opt.Val(game.PT{Value: 3}),
			StaticAbilities: []game.StaticAbility{
				game.FlyingStaticBody,
				game.LifelinkStaticBody,
			},
			OracleText: `
			Flying, lifelink
		`,
		},
	}
}
