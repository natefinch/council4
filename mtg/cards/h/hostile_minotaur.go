package h

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// HostileMinotaur is the card definition for Hostile Minotaur.
//
// Type: Creature — Minotaur
// Cost: {3}{R}
//
// Oracle text:
//
//	Haste (This creature can attack and {T} as soon as it comes under your control.)
var HostileMinotaur = newHostileMinotaur

func newHostileMinotaur() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Red),
		CardFace: game.CardFace{
			Name: "Hostile Minotaur",
			ManaCost: opt.Val(cost.Mana{
				cost.O(3),
				cost.R,
			}),
			Colors:    []color.Color{color.Red},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Minotaur},
			Power:     opt.Val(game.PT{Value: 3}),
			Toughness: opt.Val(game.PT{Value: 3}),
			StaticAbilities: []game.StaticAbility{
				game.HasteStaticBody,
			},
			OracleText: `
			Haste (This creature can attack and {T} as soon as it comes under your control.)
		`,
		},
	}
}
