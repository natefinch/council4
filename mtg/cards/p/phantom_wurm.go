package p

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// PhantomWurm is the card definition for Phantom Wurm.
//
// Type: Creature — Wurm Spirit
// Cost: {4}{G}{G}
//
// Oracle text:
//
//	This creature enters with four +1/+1 counters on it.
//	If damage would be dealt to this creature, prevent that damage. Remove a +1/+1 counter from this creature.
var PhantomWurm = newPhantomWurm

func newPhantomWurm() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Green),
		CardFace: game.CardFace{
			Name: "Phantom Wurm",
			ManaCost: opt.Val(cost.Mana{
				cost.O(4),
				cost.G,
				cost.G,
			}),
			Colors:    []color.Color{color.Green},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Wurm, types.Spirit},
			Power:     opt.Val(game.PT{Value: 2}),
			Toughness: opt.Val(game.PT{Value: 0}),
			ReplacementAbilities: []game.ReplacementAbility{
				game.EntersWithCountersReplacement("This creature enters with four +1/+1 counters on it.", game.CounterPlacement{Kind: counter.PlusOnePlusOne, Amount: 4}),
				game.DamagePreventionRemovesCounterReplacement("If damage would be dealt to this creature, prevent that damage. Remove a +1/+1 counter from this creature.", false, opt.V[game.Condition]{}),
			},
			OracleText: `
			This creature enters with four +1/+1 counters on it.
			If damage would be dealt to this creature, prevent that damage. Remove a +1/+1 counter from this creature.
		`,
		},
	}
}
