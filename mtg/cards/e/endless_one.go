package e

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// EndlessOne is the card definition for Endless One.
//
// Type: Creature — Eldrazi
// Cost: {X}
//
// Oracle text:
//
//	This creature enters with X +1/+1 counters on it.
var EndlessOne = newEndlessOne

func newEndlessOne() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name: "Endless One",
			ManaCost: opt.Val(cost.Mana{
				cost.X,
			}),
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Eldrazi},
			Power:     opt.Val(game.PT{Value: 0}),
			Toughness: opt.Val(game.PT{Value: 0}),
			ReplacementAbilities: []game.ReplacementAbility{
				game.EntersWithCountersReplacement("This creature enters with X +1/+1 counters on it.", game.CounterPlacement{Kind: counter.PlusOnePlusOne, AmountFromX: true}),
			},
			OracleText: `
			This creature enters with X +1/+1 counters on it.
		`,
		},
	}
}
