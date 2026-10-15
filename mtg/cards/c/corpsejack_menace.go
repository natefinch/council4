package c

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// CorpsejackMenace is the card definition for Corpsejack Menace.
//
// Type: Creature — Fungus
// Cost: {2}{B}{G}
//
// Oracle text:
//
//	If one or more +1/+1 counters would be put on a creature you control, twice that many +1/+1 counters are put on it instead.
var CorpsejackMenace = newCorpsejackMenace

func newCorpsejackMenace() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Black, color.Green),
		CardFace: game.CardFace{
			Name: "Corpsejack Menace",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
				cost.B,
				cost.G,
			}),
			Colors:    []color.Color{color.Black, color.Green},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Fungus},
			Power:     opt.Val(game.PT{Value: 4}),
			Toughness: opt.Val(game.PT{Value: 4}),
			ReplacementAbilities: []game.ReplacementAbility{
				game.CounterPlacementReplacement("If one or more +1/+1 counters would be put on a creature you control, twice that many +1/+1 counters are put on it instead.", 2, 0, counter.PlusOnePlusOne, game.TriggerControllerYou),
			},
			OracleText: `
			If one or more +1/+1 counters would be put on a creature you control, twice that many +1/+1 counters are put on it instead.
		`,
		},
	}
}
