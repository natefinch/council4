package p

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// PhantomNomad is the card definition for Phantom Nomad.
//
// Type: Creature — Spirit Nomad
// Cost: {1}{W}
//
// Oracle text:
//
//	This creature enters with two +1/+1 counters on it.
//	If damage would be dealt to this creature, prevent that damage. Remove a +1/+1 counter from this creature.
var PhantomNomad = newPhantomNomad

func newPhantomNomad() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White),
		CardFace: game.CardFace{
			Name: "Phantom Nomad",
			ManaCost: opt.Val(cost.Mana{
				cost.O(1),
				cost.W,
			}),
			Colors:    []color.Color{color.White},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Spirit, types.Nomad},
			Power:     opt.Val(game.PT{Value: 0}),
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
