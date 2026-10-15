package f

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// FesterhideBoar is the card definition for Festerhide Boar.
//
// Type: Creature — Boar
// Cost: {3}{G}
//
// Oracle text:
//
//	Trample
//	Morbid — This creature enters with two +1/+1 counters on it if a creature died this turn.
var FesterhideBoar = newFesterhideBoar

func newFesterhideBoar() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Green),
		CardFace: game.CardFace{
			Name: "Festerhide Boar",
			ManaCost: opt.Val(cost.Mana{
				cost.O(3),
				cost.G,
			}),
			Colors:    []color.Color{color.Green},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Boar},
			Power:     opt.Val(game.PT{Value: 3}),
			Toughness: opt.Val(game.PT{Value: 3}),
			StaticAbilities: []game.StaticAbility{
				game.TrampleStaticBody,
			},
			ReplacementAbilities: []game.ReplacementAbility{
				game.EntersWithCountersIfReplacement("Morbid — This creature enters with two +1/+1 counters on it if a creature died this turn.", &game.Condition{
					EventHistory: opt.Val(game.EventHistoryCondition{Pattern: game.TriggerPattern{
						Event:            game.EventPermanentDied,
						SubjectSelection: game.Selection{RequiredTypes: []types.Card{types.Creature}},
						StackObjectKind:  game.StackSpell,
						CounterKind:      counter.PlusOnePlusOne,
					}, Window: game.EventHistoryCurrentTurn}),
				}, game.CounterPlacement{Kind: counter.PlusOnePlusOne, Amount: 2}),
			},
			OracleText: `
			Trample
			Morbid — This creature enters with two +1/+1 counters on it if a creature died this turn.
		`,
		},
	}
}
