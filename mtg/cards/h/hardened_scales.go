package h

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// HardenedScales is the card definition for Hardened Scales.
//
// Type: Enchantment
// Cost: {G}
//
// Oracle text:
//
//	If one or more +1/+1 counters would be put on a creature you control, that many plus one +1/+1 counters are put on it instead.
var HardenedScales = newHardenedScales

func newHardenedScales() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Green),
		CardFace: game.CardFace{
			Name: "Hardened Scales",
			ManaCost: opt.Val(cost.Mana{
				cost.G,
			}),
			Colors: []color.Color{color.Green},
			Types:  []types.Card{types.Enchantment},
			ReplacementAbilities: []game.ReplacementAbility{
				game.CounterPlacementReplacement("If one or more +1/+1 counters would be put on a creature you control, that many plus one +1/+1 counters are put on it instead.", 0, 1, counter.PlusOnePlusOne, game.TriggerControllerYou),
			},
			OracleText: `
			If one or more +1/+1 counters would be put on a creature you control, that many plus one +1/+1 counters are put on it instead.
		`,
		},
	}
}
