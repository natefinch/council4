package b

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// BoonReflection is the card definition for Boon Reflection.
//
// Type: Enchantment
// Cost: {4}{W}
//
// Oracle text:
//
//	If you would gain life, you gain twice that much life instead.
var BoonReflection = newBoonReflection

func newBoonReflection() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White),
		CardFace: game.CardFace{
			Name: "Boon Reflection",
			ManaCost: opt.Val(cost.Mana{
				cost.O(4),
				cost.W,
			}),
			Colors: []color.Color{color.White},
			Types:  []types.Card{types.Enchantment},
			ReplacementAbilities: []game.ReplacementAbility{
				game.LifeGainReplacement("If you would gain life, you gain twice that much life instead.", 2, 0),
			},
			OracleText: `
			If you would gain life, you gain twice that much life instead.
		`,
		},
	}
}
