package s

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// SomberwaldSpider is the card definition for Somberwald Spider.
//
// Type: Creature — Spider
// Cost: {4}{G}
//
// Oracle text:
//
//	Reach (This creature can block creatures with flying.)
//	Morbid — This creature enters with two +1/+1 counters on it if a creature died this turn.
var SomberwaldSpider = newSomberwaldSpider

func newSomberwaldSpider() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Green),
		CardFace: game.CardFace{
			Name: "Somberwald Spider",
			ManaCost: opt.Val(cost.Mana{
				cost.O(4),
				cost.G,
			}),
			Colors:    []color.Color{color.Green},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Spider},
			Power:     opt.Val(game.PT{Value: 2}),
			Toughness: opt.Val(game.PT{Value: 4}),
			StaticAbilities: []game.StaticAbility{
				game.ReachStaticBody,
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
			Reach (This creature can block creatures with flying.)
			Morbid — This creature enters with two +1/+1 counters on it if a creature died this turn.
		`,
		},
	}
}
