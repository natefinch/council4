package g

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// GravetillerWurm is the card definition for Gravetiller Wurm.
//
// Type: Creature — Wurm
// Cost: {5}{G}
//
// Oracle text:
//
//	Trample
//	Morbid — This creature enters with four +1/+1 counters on it if a creature died this turn.
var GravetillerWurm = newGravetillerWurm

func newGravetillerWurm() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Green),
		CardFace: game.CardFace{
			Name: "Gravetiller Wurm",
			ManaCost: opt.Val(cost.Mana{
				cost.O(5),
				cost.G,
			}),
			Colors:    []color.Color{color.Green},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Wurm},
			Power:     opt.Val(game.PT{Value: 4}),
			Toughness: opt.Val(game.PT{Value: 4}),
			StaticAbilities: []game.StaticAbility{
				game.TrampleStaticBody,
			},
			ReplacementAbilities: []game.ReplacementAbility{
				game.EntersWithCountersIfReplacement("Morbid — This creature enters with four +1/+1 counters on it if a creature died this turn.", &game.Condition{
					EventHistory: opt.Val(game.EventHistoryCondition{Pattern: game.TriggerPattern{
						Event:            game.EventPermanentDied,
						SubjectSelection: game.Selection{RequiredTypes: []types.Card{types.Creature}},
						StackObjectKind:  game.StackSpell,
						CounterKind:      counter.PlusOnePlusOne,
					}, Window: game.EventHistoryCurrentTurn}),
				}, game.CounterPlacement{Kind: counter.PlusOnePlusOne, Amount: 4}),
			},
			OracleText: `
			Trample
			Morbid — This creature enters with four +1/+1 counters on it if a creature died this turn.
		`,
		},
	}
}
