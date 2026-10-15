package p

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// PrimalVigor is the card definition for Primal Vigor.
//
// Type: Enchantment
// Cost: {4}{G}
//
// Oracle text:
//
//	If one or more tokens would be created, twice that many of those tokens are created instead.
//	If one or more +1/+1 counters would be put on a creature, twice that many +1/+1 counters are put on that creature instead.
var PrimalVigor = newPrimalVigor

func newPrimalVigor() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Green),
		CardFace: game.CardFace{
			Name: "Primal Vigor",
			ManaCost: opt.Val(cost.Mana{
				cost.O(4),
				cost.G,
			}),
			Colors: []color.Color{color.Green},
			Types:  []types.Card{types.Enchantment},
			ReplacementAbilities: []game.ReplacementAbility{
				game.TokenCreationReplacementFiltered("If one or more tokens would be created, twice that many of those tokens are created instead.", &game.TokenCreationReplacementSpec{Multiplier: 2, Filter: game.TriggerControllerAny}),
				game.CounterPlacementReplacement("If one or more +1/+1 counters would be put on a creature, twice that many +1/+1 counters are put on that creature instead.", 2, 0, counter.PlusOnePlusOne, game.TriggerControllerAny),
			},
			OracleText: `
			If one or more tokens would be created, twice that many of those tokens are created instead.
			If one or more +1/+1 counters would be put on a creature, twice that many +1/+1 counters are put on that creature instead.
		`,
		},
	}
}
