package p

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// PhantomNantuko is the card definition for Phantom Nantuko.
//
// Type: Creature — Insect Spirit
// Cost: {2}{G}
//
// Oracle text:
//
//	Trample
//	This creature enters with two +1/+1 counters on it.
//	If damage would be dealt to this creature, prevent that damage. Remove a +1/+1 counter from this creature.
//	{T}: Put a +1/+1 counter on this creature.
var PhantomNantuko = newPhantomNantuko

func newPhantomNantuko() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Green),
		CardFace: game.CardFace{
			Name: "Phantom Nantuko",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
				cost.G,
			}),
			Colors:    []color.Color{color.Green},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Insect, types.Spirit},
			Power:     opt.Val(game.PT{Value: 0}),
			Toughness: opt.Val(game.PT{Value: 0}),
			StaticAbilities: []game.StaticAbility{
				game.TrampleStaticBody,
			},
			ActivatedAbilities: []game.ActivatedAbility{
				game.ActivatedAbility{
					Text: "{T}: Put a +1/+1 counter on this creature.",
					AdditionalCosts: []cost.Additional{
						cost.Additional{
							Kind:        cost.AdditionalTap,
							CounterKind: counter.PlusOnePlusOne,
						},
					},
					ZoneOfFunction: zone.Battlefield,
					Content: game.AbilityContent{
						Modes: []game.Mode{
							game.Mode{
								Sequence: []game.Instruction{
									game.Instruction{
										Primitive: game.AddCounter{
											Amount:      game.Fixed(1),
											Object:      game.SourcePermanentReference(),
											CounterKind: counter.PlusOnePlusOne,
										},
									},
								},
							},
						},
						MinModes: 1,
						MaxModes: 1,
					},
				},
			},
			ReplacementAbilities: []game.ReplacementAbility{
				game.EntersWithCountersReplacement("This creature enters with two +1/+1 counters on it.", game.CounterPlacement{Kind: counter.PlusOnePlusOne, Amount: 2}),
				game.DamagePreventionRemovesCounterReplacement("If damage would be dealt to this creature, prevent that damage. Remove a +1/+1 counter from this creature.", false, opt.V[game.Condition]{}),
			},
			OracleText: `
			Trample
			This creature enters with two +1/+1 counters on it.
			If damage would be dealt to this creature, prevent that damage. Remove a +1/+1 counter from this creature.
			{T}: Put a +1/+1 counter on this creature.
		`,
		},
	}
}
