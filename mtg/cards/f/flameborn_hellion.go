package f

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// FlamebornHellion is the card definition for Flameborn Hellion.
//
// Type: Creature — Hellion
// Cost: {5}{R}
//
// Oracle text:
//
//	Haste
//	This creature attacks each combat if able.
var FlamebornHellion = newFlamebornHellion

func newFlamebornHellion() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Red),
		CardFace: game.CardFace{
			Name: "Flameborn Hellion",
			ManaCost: opt.Val(cost.Mana{
				cost.O(5),
				cost.R,
			}),
			Colors:    []color.Color{color.Red},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Hellion},
			Power:     opt.Val(game.PT{Value: 5}),
			Toughness: opt.Val(game.PT{Value: 4}),
			StaticAbilities: []game.StaticAbility{
				game.HasteStaticBody,
				game.MustAttackStaticBody,
			},
			OracleText: `
			Haste
			This creature attacks each combat if able.
		`,
		},
	}
}
