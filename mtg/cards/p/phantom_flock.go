package p

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// PhantomFlock is the card definition for Phantom Flock.
//
// Type: Creature — Bird Soldier Spirit
// Cost: {3}{W}{W}
//
// Oracle text:
//
//	Flying
//	This creature enters with three +1/+1 counters on it.
//	If damage would be dealt to this creature, prevent that damage. Remove a +1/+1 counter from this creature.
var PhantomFlock = newPhantomFlock

func newPhantomFlock() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White),
		CardFace: game.CardFace{
			Name: "Phantom Flock",
			ManaCost: opt.Val(cost.Mana{
				cost.O(3),
				cost.W,
				cost.W,
			}),
			Colors:    []color.Color{color.White},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Bird, types.Soldier, types.Spirit},
			Power:     opt.Val(game.PT{Value: 0}),
			Toughness: opt.Val(game.PT{Value: 0}),
			StaticAbilities: []game.StaticAbility{
				game.FlyingStaticBody,
			},
			ReplacementAbilities: []game.ReplacementAbility{
				game.EntersWithCountersReplacement("This creature enters with three +1/+1 counters on it.", game.CounterPlacement{Kind: counter.PlusOnePlusOne, Amount: 3}),
				game.DamagePreventionRemovesCounterReplacement("If damage would be dealt to this creature, prevent that damage. Remove a +1/+1 counter from this creature.", false, opt.V[game.Condition]{}),
			},
			OracleText: `
			Flying
			This creature enters with three +1/+1 counters on it.
			If damage would be dealt to this creature, prevent that damage. Remove a +1/+1 counter from this creature.
		`,
		},
	}
}
