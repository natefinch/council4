package c

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// CopyEnchantment is the card definition for Copy Enchantment.
//
// Type: Enchantment
// Cost: {2}{U}
//
// Oracle text:
//
//	You may have this enchantment enter as a copy of any enchantment on the battlefield.
var CopyEnchantment = newCopyEnchantment

func newCopyEnchantment() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue),
		CardFace: game.CardFace{
			Name: "Copy Enchantment",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
				cost.U,
			}),
			Colors: []color.Color{color.Blue},
			Types:  []types.Card{types.Enchantment},
			ReplacementAbilities: []game.ReplacementAbility{
				game.EntersAsCopyReplacement("You may have this enchantment enter as a copy of any enchantment on the battlefield.", &game.Selection{RequiredTypes: []types.Card{types.Enchantment}}, true, false, nil, false, nil, nil),
			},
			OracleText: `
			You may have this enchantment enter as a copy of any enchantment on the battlefield.
		`,
		},
	}
}
