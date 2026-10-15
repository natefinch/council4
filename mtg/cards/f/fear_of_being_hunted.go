package f

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// FearOfBeingHunted is the card definition for Fear of Being Hunted.
//
// Type: Enchantment Creature — Nightmare
// Cost: {1}{R}{R}
//
// Oracle text:
//
//	Haste
//	This creature must be blocked if able.
var FearOfBeingHunted = newFearOfBeingHunted

func newFearOfBeingHunted() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Red),
		CardFace: game.CardFace{
			Name: "Fear of Being Hunted",
			ManaCost: opt.Val(cost.Mana{
				cost.O(1),
				cost.R,
				cost.R,
			}),
			Colors:    []color.Color{color.Red},
			Types:     []types.Card{types.Enchantment, types.Creature},
			Subtypes:  []types.Sub{types.Nightmare},
			Power:     opt.Val(game.PT{Value: 4}),
			Toughness: opt.Val(game.PT{Value: 2}),
			StaticAbilities: []game.StaticAbility{
				game.HasteStaticBody,
				game.MustBeBlockedStaticBody,
			},
			OracleText: `
			Haste
			This creature must be blocked if able.
		`,
		},
	}
}
