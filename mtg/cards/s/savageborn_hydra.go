package s

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// SavagebornHydra is the card definition for Savageborn Hydra.
//
// Type: Creature — Hydra
// Cost: {X}{R}{G}
//
// Oracle text:
//
//	Double strike
//	This creature enters with X +1/+1 counters on it.
//	{1}{R/G}: Put a +1/+1 counter on this creature. Activate only as a sorcery.
var SavagebornHydra = newSavagebornHydra

func newSavagebornHydra() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Red, color.Green),
		CardFace: game.CardFace{
			Name: "Savageborn Hydra",
			ManaCost: opt.Val(cost.Mana{
				cost.X,
				cost.R,
				cost.G,
			}),
			Colors:    []color.Color{color.Green, color.Red},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Hydra},
			Power:     opt.Val(game.PT{Value: 0}),
			Toughness: opt.Val(game.PT{Value: 0}),
			StaticAbilities: []game.StaticAbility{
				game.DoubleStrikeStaticBody,
			},
			ActivatedAbilities: []game.ActivatedAbility{
				game.ActivatedAbility{
					Text: "{1}{R/G}: Put a +1/+1 counter on this creature. Activate only as a sorcery.",
					ManaCost: opt.Val(cost.Mana{
						cost.O(1),
						cost.HybridMana(mana.R, mana.G),
					}),
					ZoneOfFunction: zone.Battlefield,
					Timing:         game.SorceryOnly,
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
				game.EntersWithCountersReplacement("This creature enters with X +1/+1 counters on it.", game.CounterPlacement{Kind: counter.PlusOnePlusOne, AmountFromX: true}),
			},
			OracleText: `
			Double strike
			This creature enters with X +1/+1 counters on it.
			{1}{R/G}: Put a +1/+1 counter on this creature. Activate only as a sorcery.
		`,
		},
	}
}
