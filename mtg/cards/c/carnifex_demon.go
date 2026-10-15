package c

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// CarnifexDemon is the card definition for Carnifex Demon.
//
// Type: Creature — Phyrexian Demon
// Cost: {4}{B}{B}
//
// Oracle text:
//
//	Flying
//	This creature enters with two -1/-1 counters on it.
//	{B}, Remove a -1/-1 counter from this creature: Put a -1/-1 counter on each other creature.
var CarnifexDemon = newCarnifexDemon

func newCarnifexDemon() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Black),
		CardFace: game.CardFace{
			Name: "Carnifex Demon",
			ManaCost: opt.Val(cost.Mana{
				cost.O(4),
				cost.B,
				cost.B,
			}),
			Colors:    []color.Color{color.Black},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Phyrexian, types.Demon},
			Power:     opt.Val(game.PT{Value: 6}),
			Toughness: opt.Val(game.PT{Value: 6}),
			StaticAbilities: []game.StaticAbility{
				game.FlyingStaticBody,
			},
			ActivatedAbilities: []game.ActivatedAbility{
				game.ActivatedAbility{
					Text: "{B}, Remove a -1/-1 counter from this creature: Put a -1/-1 counter on each other creature.",
					ManaCost: opt.Val(cost.Mana{
						cost.B,
					}),
					AdditionalCosts: []cost.Additional{
						cost.Additional{
							Kind:        cost.AdditionalRemoveCounter,
							Text:        "Remove a -1/-1 counter from this creature",
							Amount:      1,
							CounterKind: counter.MinusOneMinusOne,
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
											Group:       game.BattlefieldGroupExcluding(game.Selection{RequiredTypes: []types.Card{types.Creature}}, game.SourcePermanentReference()),
											CounterKind: counter.MinusOneMinusOne,
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
				game.EntersWithCountersReplacement("This creature enters with two -1/-1 counters on it.", game.CounterPlacement{Kind: counter.MinusOneMinusOne, Amount: 2}),
			},
			OracleText: `
			Flying
			This creature enters with two -1/-1 counters on it.
			{B}, Remove a -1/-1 counter from this creature: Put a -1/-1 counter on each other creature.
		`,
		},
	}
}
