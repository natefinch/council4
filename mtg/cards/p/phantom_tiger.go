package p

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// PhantomTiger is the card definition for Phantom Tiger.
//
// Type: Creature — Cat Spirit
// Cost: {2}{G}
//
// Oracle text:
//
//	This creature enters with two +1/+1 counters on it.
//	If damage would be dealt to this creature, prevent that damage. Remove a +1/+1 counter from this creature.
var PhantomTiger = newPhantomTiger

func newPhantomTiger() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Green),
		CardFace: game.CardFace{
			Name: "Phantom Tiger",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
				cost.G,
			}),
			Colors:    []color.Color{color.Green},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Cat, types.Spirit},
			Power:     opt.Val(game.PT{Value: 1}),
			Toughness: opt.Val(game.PT{Value: 0}),
			ReplacementAbilities: []game.ReplacementAbility{
				game.EntersWithCountersReplacement("This creature enters with two +1/+1 counters on it.", game.CounterPlacement{Kind: counter.PlusOnePlusOne, Amount: 2}),
				game.DamagePreventionRemovesCounterReplacement("If damage would be dealt to this creature, prevent that damage. Remove a +1/+1 counter from this creature.", false, opt.V[game.Condition]{}),
			},
			OracleText: `
			This creature enters with two +1/+1 counters on it.
			If damage would be dealt to this creature, prevent that damage. Remove a +1/+1 counter from this creature.
		`,
		},
	}
}
