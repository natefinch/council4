package e

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// EffortlessMaster is the card definition for Effortless Master.
//
// Type: Creature — Orc Monk
// Cost: {2}{U}{R}
//
// Oracle text:
//
//	Vigilance
//	Menace (This creature can't be blocked except by two or more creatures.)
//	This creature enters with two +1/+1 counters on it if you've cast two or more spells this turn.
var EffortlessMaster = newEffortlessMaster

func newEffortlessMaster() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue, color.Red),
		CardFace: game.CardFace{
			Name: "Effortless Master",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
				cost.U,
				cost.R,
			}),
			Colors:    []color.Color{color.Red, color.Blue},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Orc, types.Monk},
			Power:     opt.Val(game.PT{Value: 4}),
			Toughness: opt.Val(game.PT{Value: 3}),
			StaticAbilities: []game.StaticAbility{
				game.VigilanceStaticBody,
				game.MenaceStaticBody,
			},
			ReplacementAbilities: []game.ReplacementAbility{
				game.EntersWithCountersIfReplacement("This creature enters with two +1/+1 counters on it if you've cast two or more spells this turn.", &game.Condition{
					EventHistory: opt.Val(game.EventHistoryCondition{Pattern: game.TriggerPattern{
						Event:           game.EventSpellCast,
						Controller:      game.TriggerControllerYou,
						StackObjectKind: game.StackSpell,
						CounterKind:     counter.PlusOnePlusOne,
					}, Window: game.EventHistoryCurrentTurn, MinCount: 2}),
				}, game.CounterPlacement{Kind: counter.PlusOnePlusOne, Amount: 2}),
			},
			OracleText: `
			Vigilance
			Menace (This creature can't be blocked except by two or more creatures.)
			This creature enters with two +1/+1 counters on it if you've cast two or more spells this turn.
		`,
		},
	}
}
