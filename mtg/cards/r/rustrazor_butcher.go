package r

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// RustrazorButcher is the card definition for Rustrazor Butcher.
//
// Type: Creature — Goblin Warrior
// Cost: {1}{R}
//
// Oracle text:
//
//	First strike
//	Wither (This deals damage to creatures in the form of -1/-1 counters.)
var RustrazorButcher = newRustrazorButcher

func newRustrazorButcher() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Red),
		CardFace: game.CardFace{
			Name: "Rustrazor Butcher",
			ManaCost: opt.Val(cost.Mana{
				cost.O(1),
				cost.R,
			}),
			Colors:    []color.Color{color.Red},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Goblin, types.Warrior},
			Power:     opt.Val(game.PT{Value: 1}),
			Toughness: opt.Val(game.PT{Value: 2}),
			StaticAbilities: []game.StaticAbility{
				game.FirstStrikeStaticBody,
				game.WitherStaticBody,
			},
			OracleText: `
			First strike
			Wither (This deals damage to creatures in the form of -1/-1 counters.)
		`,
		},
	}
}
