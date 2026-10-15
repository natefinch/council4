package s

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// SkatewingSpy is the card definition for Skatewing Spy.
//
// Type: Creature — Vedalken Rogue Mutant
// Cost: {3}{U}
//
// Oracle text:
//
//	{5}{U}: Adapt 2. (If this creature has no +1/+1 counters on it, put two +1/+1 counters on it.)
//	Each creature you control with a +1/+1 counter on it has flying.
var SkatewingSpy = newSkatewingSpy

func newSkatewingSpy() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue),
		CardFace: game.CardFace{
			Name: "Skatewing Spy",
			ManaCost: opt.Val(cost.Mana{
				cost.O(3),
				cost.U,
			}),
			Colors:    []color.Color{color.Blue},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Vedalken, types.Rogue, types.Mutant},
			Power:     opt.Val(game.PT{Value: 2}),
			Toughness: opt.Val(game.PT{Value: 3}),
			StaticAbilities: []game.StaticAbility{
				game.StaticAbility{
					Text: "Each creature you control with a +1/+1 counter on it has flying.",
					ContinuousEffects: []game.ContinuousEffect{
						game.ContinuousEffect{
							Duration:    game.DurationPermanent,
							Group:       game.ObjectControlledGroup(game.SourcePermanentReference(), game.Selection{RequiredTypes: []types.Card{types.Creature}, MatchCounter: true, RequiredCounter: counter.PlusOnePlusOne}),
							Layer:       game.LayerAbility,
							AddKeywords: []game.Keyword{game.Flying},
						},
					},
				},
			},
			ActivatedAbilities: []game.ActivatedAbility{
				game.ActivatedAbility{
					Text: "{5}{U}: Adapt 2. (If this creature has no +1/+1 counters on it, put two +1/+1 counters on it.)",
					ManaCost: opt.Val(cost.Mana{
						cost.O(5),
						cost.U,
					}),
					ZoneOfFunction: zone.Battlefield,
					Content: game.AbilityContent{
						Modes: []game.Mode{
							game.Mode{
								Sequence: []game.Instruction{
									game.Instruction{
										Primitive: game.Adapt{
											Object: game.SourcePermanentReference(),
											Amount: game.Fixed(2),
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
			OracleText: `
			{5}{U}: Adapt 2. (If this creature has no +1/+1 counters on it, put two +1/+1 counters on it.)
			Each creature you control with a +1/+1 counter on it has flying.
		`,
		},
	}
}
