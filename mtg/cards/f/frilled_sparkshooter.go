package f

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// FrilledSparkshooter is the card definition for Frilled Sparkshooter.
//
// Type: Creature — Lizard Archer
// Cost: {3}{R}
//
// Oracle text:
//
//	Reach, menace
//	This creature enters with a +1/+1 counter on it if an opponent lost life this turn.
var FrilledSparkshooter = newFrilledSparkshooter

func newFrilledSparkshooter() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Red),
		CardFace: game.CardFace{
			Name: "Frilled Sparkshooter",
			ManaCost: opt.Val(cost.Mana{
				cost.O(3),
				cost.R,
			}),
			Colors:    []color.Color{color.Red},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Lizard, types.Archer},
			Power:     opt.Val(game.PT{Value: 3}),
			Toughness: opt.Val(game.PT{Value: 3}),
			StaticAbilities: []game.StaticAbility{
				game.ReachStaticBody,
				game.MenaceStaticBody,
			},
			ReplacementAbilities: []game.ReplacementAbility{
				game.EntersWithCountersIfReplacement("This creature enters with a +1/+1 counter on it if an opponent lost life this turn.", &game.Condition{
					EventHistory: opt.Val(game.EventHistoryCondition{Pattern: game.TriggerPattern{
						Event:           game.EventLifeLost,
						Player:          game.TriggerPlayerOpponent,
						StackObjectKind: game.StackSpell,
						CounterKind:     counter.PlusOnePlusOne,
					}, Window: game.EventHistoryCurrentTurn}),
				}, game.CounterPlacement{Kind: counter.PlusOnePlusOne, Amount: 1}),
			},
			OracleText: `
			Reach, menace
			This creature enters with a +1/+1 counter on it if an opponent lost life this turn.
		`,
		},
	}
}
