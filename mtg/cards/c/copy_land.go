package c

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// CopyLand is the card definition for Copy Land.
//
// Type: Enchantment
// Cost: {2}{U}
//
// Oracle text:
//
//	You may have this enchantment enter as a copy of any land on the battlefield, except it's an enchantment in addition to its other types.
var CopyLand = newCopyLand

func newCopyLand() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue),
		CardFace: game.CardFace{
			Name: "Copy Land",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
				cost.U,
			}),
			Colors: []color.Color{color.Blue},
			Types:  []types.Card{types.Enchantment},
			ReplacementAbilities: []game.ReplacementAbility{
				game.EntersAsCopyReplacement("You may have this enchantment enter as a copy of any land on the battlefield, except it's an enchantment in addition to its other types.", &game.Selection{RequiredTypes: []types.Card{types.Land}}, true, false, nil, false, nil, nil, types.Enchantment),
			},
			OracleText: `
			You may have this enchantment enter as a copy of any land on the battlefield, except it's an enchantment in addition to its other types.
		`,
		},
	}
}
