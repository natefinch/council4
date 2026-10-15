package u

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// UtvaraScalper is the card definition for Utvara Scalper.
//
// Type: Creature — Goblin Scout
// Cost: {1}{R}
//
// Oracle text:
//
//	Flying
//	This creature attacks each combat if able.
var UtvaraScalper = newUtvaraScalper

func newUtvaraScalper() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Red),
		CardFace: game.CardFace{
			Name: "Utvara Scalper",
			ManaCost: opt.Val(cost.Mana{
				cost.O(1),
				cost.R,
			}),
			Colors:    []color.Color{color.Red},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Goblin, types.Scout},
			Power:     opt.Val(game.PT{Value: 1}),
			Toughness: opt.Val(game.PT{Value: 2}),
			StaticAbilities: []game.StaticAbility{
				game.FlyingStaticBody,
				game.MustAttackStaticBody,
			},
			OracleText: `
			Flying
			This creature attacks each combat if able.
		`,
		},
	}
}
