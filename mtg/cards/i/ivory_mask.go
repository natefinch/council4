package i

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// IvoryMask is the card definition for Ivory Mask.
//
// Type: Enchantment
// Cost: {2}{W}{W}
//
// Oracle text:
//
//	You have shroud. (You can't be the target of spells or abilities.)
var IvoryMask = newIvoryMask()

func newIvoryMask() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White),
		CardFace: game.CardFace{
			Name: "Ivory Mask",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
				cost.W,
				cost.W,
			}),
			Colors: []color.Color{color.White},
			Types:  []types.Card{types.Enchantment},
			StaticAbilities: []game.StaticAbility{
				game.PlayerShroudStaticBody,
			},
			OracleText: `
			You have shroud. (You can't be the target of spells or abilities.)
		`,
		},
	}
}
