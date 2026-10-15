package a

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// AnointedProcession is the card definition for Anointed Procession.
//
// Type: Enchantment
// Cost: {3}{W}
//
// Oracle text:
//
//	If an effect would create one or more tokens under your control, it creates twice that many of those tokens instead.
var AnointedProcession = newAnointedProcession

func newAnointedProcession() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White),
		CardFace: game.CardFace{
			Name: "Anointed Procession",
			ManaCost: opt.Val(cost.Mana{
				cost.O(3),
				cost.W,
			}),
			Colors: []color.Color{color.White},
			Types:  []types.Card{types.Enchantment},
			ReplacementAbilities: []game.ReplacementAbility{
				game.TokenCreationReplacement("If an effect would create one or more tokens under your control, it creates twice that many of those tokens instead.", 2, game.TriggerControllerYou),
			},
			OracleText: `
			If an effect would create one or more tokens under your control, it creates twice that many of those tokens instead.
		`,
		},
	}
}
