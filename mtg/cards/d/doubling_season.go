package d

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// DoublingSeason is the card definition for Doubling Season.
//
// Type: Enchantment
// Cost: {4}{G}
//
// Oracle text:
//
//	If an effect would create one or more tokens under your control, it creates twice that many of those tokens instead.
//	If an effect would put one or more counters on a permanent you control, it puts twice that many of those counters on that permanent instead.
var DoublingSeason = newDoublingSeason

func newDoublingSeason() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Green),
		CardFace: game.CardFace{
			Name: "Doubling Season",
			ManaCost: opt.Val(cost.Mana{
				cost.O(4),
				cost.G,
			}),
			Colors: []color.Color{color.Green},
			Types:  []types.Card{types.Enchantment},
			ReplacementAbilities: []game.ReplacementAbility{
				game.TokenCreationReplacement("If an effect would create one or more tokens under your control, it creates twice that many of those tokens instead.", 2, game.TriggerControllerYou),
				game.ControlledPermanentCounterPlacementReplacement("If an effect would put one or more counters on a permanent you control, it puts twice that many of those counters on that permanent instead.", 2, 0, game.TriggerControllerYou),
			},
			OracleText: `
			If an effect would create one or more tokens under your control, it creates twice that many of those tokens instead.
			If an effect would put one or more counters on a permanent you control, it puts twice that many of those counters on that permanent instead.
		`,
		},
	}
}
