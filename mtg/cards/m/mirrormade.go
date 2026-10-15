package m

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// Mirrormade is the card definition for Mirrormade.
//
// Type: Enchantment
// Cost: {1}{U}{U}
//
// Oracle text:
//
//	You may have this enchantment enter as a copy of any artifact or enchantment on the battlefield.
var Mirrormade = newMirrormade

func newMirrormade() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue),
		CardFace: game.CardFace{
			Name: "Mirrormade",
			ManaCost: opt.Val(cost.Mana{
				cost.O(1),
				cost.U,
				cost.U,
			}),
			Colors: []color.Color{color.Blue},
			Types:  []types.Card{types.Enchantment},
			ReplacementAbilities: []game.ReplacementAbility{
				game.EntersAsCopyReplacement("You may have this enchantment enter as a copy of any artifact or enchantment on the battlefield.", &game.Selection{RequiredTypesAny: []types.Card{types.Artifact, types.Enchantment}}, true, false, nil, false, nil, nil),
			},
			OracleText: `
			You may have this enchantment enter as a copy of any artifact or enchantment on the battlefield.
		`,
		},
	}
}
