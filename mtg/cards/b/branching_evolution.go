package b

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// BranchingEvolution is the card definition for Branching Evolution.
//
// Type: Enchantment
// Cost: {2}{G}
//
// Oracle text:
//
//	If one or more +1/+1 counters would be put on a creature you control, twice that many +1/+1 counters are put on that creature instead.
var BranchingEvolution = newBranchingEvolution

func newBranchingEvolution() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Green),
		CardFace: game.CardFace{
			Name: "Branching Evolution",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
				cost.G,
			}),
			Colors: []color.Color{color.Green},
			Types:  []types.Card{types.Enchantment},
			ReplacementAbilities: []game.ReplacementAbility{
				game.CounterPlacementReplacement("If one or more +1/+1 counters would be put on a creature you control, twice that many +1/+1 counters are put on that creature instead.", 2, 0, counter.PlusOnePlusOne, game.TriggerControllerYou),
			},
			OracleText: `
			If one or more +1/+1 counters would be put on a creature you control, twice that many +1/+1 counters are put on that creature instead.
		`,
		},
	}
}
