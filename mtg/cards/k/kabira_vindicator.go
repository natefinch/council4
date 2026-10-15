package k

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// KabiraVindicator is the card definition for Kabira Vindicator.
//
// Type: Creature — Human Knight
// Cost: {3}{W}
//
// Oracle text:
//
//	Level up {2}{W} ({2}{W}: Put a level counter on this. Level up only as a sorcery.)
//	LEVEL 2-4
//	3/6
//	Other creatures you control get +1/+1.
//	LEVEL 5+
//	4/8
//	Other creatures you control get +2/+2.
var KabiraVindicator = newKabiraVindicator

func newKabiraVindicator() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White),
		CardFace: game.CardFace{
			Name: "Kabira Vindicator",
			ManaCost: opt.Val(cost.Mana{
				cost.O(3),
				cost.W,
			}),
			Colors:    []color.Color{color.White},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Human, types.Knight},
			Power:     opt.Val(game.PT{Value: 2}),
			Toughness: opt.Val(game.PT{Value: 4}),
			StaticAbilities: []game.StaticAbility{
				game.StaticAbility{
					Text: "LEVEL 2-4\n3/6",
					Condition: opt.Val(game.Condition{
						SourceLevelCountersAtLeast:  2,
						SourceLevelCountersLessThan: 5,
						SourceCounterKind:           counter.PlusOnePlusOne,
					}),
					ContinuousEffects: []game.ContinuousEffect{
						game.ContinuousEffect{
							Duration:       game.DurationPermanent,
							AffectedSource: true,
							Layer:          game.LayerPowerToughnessSet,
							SetPower: opt.Val(game.PT{
								Value: 3,
							}),
							SetToughness: opt.Val(game.PT{
								Value: 6,
							}),
						},
					},
				},
				game.StaticAbility{
					Text: "Other creatures you control get +1/+1.",
					Condition: opt.Val(game.Condition{
						SourceLevelCountersAtLeast:  2,
						SourceLevelCountersLessThan: 5,
						SourceCounterKind:           counter.PlusOnePlusOne,
					}),
					ContinuousEffects: []game.ContinuousEffect{
						game.ContinuousEffect{
							Duration:       game.DurationPermanent,
							Group:          game.ObjectControlledGroupExcluding(game.SourcePermanentReference(), game.Selection{RequiredTypes: []types.Card{types.Creature}}, game.SourcePermanentReference()),
							Layer:          game.LayerPowerToughnessModify,
							PowerDelta:     1,
							ToughnessDelta: 1,
						},
					},
				},
				game.StaticAbility{
					Text: "LEVEL 5+\n4/8",
					Condition: opt.Val(game.Condition{
						SourceLevelCountersAtLeast: 5,
						SourceCounterKind:          counter.PlusOnePlusOne,
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
								Value: 8,
							}),
						},
					},
				},
				game.StaticAbility{
					Text: "Other creatures you control get +2/+2.",
					Condition: opt.Val(game.Condition{
						SourceLevelCountersAtLeast: 5,
						SourceCounterKind:          counter.PlusOnePlusOne,
					}),
					ContinuousEffects: []game.ContinuousEffect{
						game.ContinuousEffect{
							Duration:       game.DurationPermanent,
							Group:          game.ObjectControlledGroupExcluding(game.SourcePermanentReference(), game.Selection{RequiredTypes: []types.Card{types.Creature}}, game.SourcePermanentReference()),
							Layer:          game.LayerPowerToughnessModify,
							PowerDelta:     2,
							ToughnessDelta: 2,
						},
					},
				},
			},
			ActivatedAbilities: []game.ActivatedAbility{
				game.ActivatedAbility{
					Text: "Level up {2}{W} ({2}{W}: Put a level counter on this. Level up only as a sorcery.)",
					ManaCost: opt.Val(cost.Mana{
						cost.O(2),
						cost.W,
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
			Level up {2}{W} ({2}{W}: Put a level counter on this. Level up only as a sorcery.)
			LEVEL 2-4
			3/6
			Other creatures you control get +1/+1.
			LEVEL 5+
			4/8
			Other creatures you control get +2/+2.
		`,
		},
	}
}
