package h

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// Hexdrinker is the card definition for Hexdrinker.
//
// Type: Creature — Snake
// Cost: {G}
//
// Oracle text:
//
//	Level up {1} ({1}: Put a level counter on this. Level up only as a sorcery.)
//	LEVEL 3-7
//	4/4
//	Protection from instants
//	LEVEL 8+
//	6/6
//	Protection from everything
var Hexdrinker = newHexdrinker

func newHexdrinker() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Green),
		CardFace: game.CardFace{
			Name: "Hexdrinker",
			ManaCost: opt.Val(cost.Mana{
				cost.G,
			}),
			Colors:    []color.Color{color.Green},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Snake},
			Power:     opt.Val(game.PT{Value: 2}),
			Toughness: opt.Val(game.PT{Value: 1}),
			StaticAbilities: []game.StaticAbility{
				game.StaticAbility{
					Text: "LEVEL 3-7\n4/4",
					Condition: opt.Val(game.Condition{
						SourceLevelCountersAtLeast:  3,
						SourceLevelCountersLessThan: 8,
						SourceCounterKind:           counter.PlusOnePlusOne,
					}),
					ContinuousEffects: []game.ContinuousEffect{
						game.ContinuousEffect{
							Duration:       game.DurationPermanent,
							AffectedSource: true,
							Layer:          game.LayerPowerToughnessSet,
							SetPower: opt.Val(game.PT{
								Value: 4,
							}),
							SetToughness: opt.Val(game.PT{
								Value: 4,
							}),
						},
					},
				},
				game.StaticAbility{
					Text: "Protection from instants",
					Condition: opt.Val(game.Condition{
						SourceLevelCountersAtLeast:  3,
						SourceLevelCountersLessThan: 8,
						SourceCounterKind:           counter.PlusOnePlusOne,
					}),
					KeywordAbilities: []game.KeywordAbility{
						game.ProtectionKeyword{
							FromTypes: []types.Card{types.Instant},
						},
					},
				},
				game.StaticAbility{
					Text: "LEVEL 8+\n6/6",
					Condition: opt.Val(game.Condition{
						SourceLevelCountersAtLeast: 8,
						SourceCounterKind:          counter.PlusOnePlusOne,
					}),
					ContinuousEffects: []game.ContinuousEffect{
						game.ContinuousEffect{
							Duration:       game.DurationPermanent,
							AffectedSource: true,
							Layer:          game.LayerPowerToughnessSet,
							SetPower: opt.Val(game.PT{
								Value: 6,
							}),
							SetToughness: opt.Val(game.PT{
								Value: 6,
							}),
						},
					},
				},
				game.StaticAbility{
					Text: "Protection from everything",
					Condition: opt.Val(game.Condition{
						SourceLevelCountersAtLeast: 8,
						SourceCounterKind:          counter.PlusOnePlusOne,
					}),
					KeywordAbilities: []game.KeywordAbility{
						game.ProtectionKeyword{
							Everything: true,
						},
					},
				},
			},
			ActivatedAbilities: []game.ActivatedAbility{
				game.ActivatedAbility{
					Text: "Level up {1} ({1}: Put a level counter on this. Level up only as a sorcery.)",
					ManaCost: opt.Val(cost.Mana{
						cost.O(1),
					}),
					Timing: game.SorceryOnly,
					Content: game.AbilityContent{
						Modes: []game.Mode{
							game.Mode{
								Sequence: []game.Instruction{
									game.Instruction{
										Primitive: game.AddCounter{
											Amount:      game.Fixed(1),
											Object:      game.SourcePermanentReference(),
											CounterKind: counter.Level,
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
			Level up {1} ({1}: Put a level counter on this. Level up only as a sorcery.)
			LEVEL 3-7
			4/4
			Protection from instants
			LEVEL 8+
			6/6
			Protection from everything
		`,
		},
	}
}
