package c

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// CoralhelmCommander is the card definition for Coralhelm Commander.
//
// Type: Creature — Merfolk Soldier
// Cost: {U}{U}
//
// Oracle text:
//
//	Level up {1} ({1}: Put a level counter on this. Level up only as a sorcery.)
//	LEVEL 2-3
//	3/3
//	Flying
//	LEVEL 4+
//	4/4
//	Flying
//	Other Merfolk creatures you control get +1/+1.
var CoralhelmCommander = newCoralhelmCommander()

func newCoralhelmCommander() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue),
		CardFace: game.CardFace{
			Name: "Coralhelm Commander",
			ManaCost: opt.Val(cost.Mana{
				cost.U,
				cost.U,
			}),
			Colors:    []color.Color{color.Blue},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Merfolk, types.Soldier},
			Power:     opt.Val(game.PT{Value: 2}),
			Toughness: opt.Val(game.PT{Value: 2}),
			StaticAbilities: []game.StaticAbility{
				game.StaticAbility{
					Condition: opt.Val(game.Condition{
						SourceLevelCountersAtLeast:  2,
						SourceLevelCountersLessThan: 4,
					}),
					ContinuousEffects: []game.ContinuousEffect{
						game.ContinuousEffect{
							Layer:          game.LayerPowerToughnessSet,
							AffectedSource: true,
							SetPower:       opt.Val(game.PT{Value: 3}),
							SetToughness:   opt.Val(game.PT{Value: 3}),
						},
					},
				},
				game.StaticAbility{
					KeywordAbilities: []game.KeywordAbility{
						game.SimpleKeyword{Kind: game.Flying},
					},
					Condition: opt.Val(game.Condition{
						SourceLevelCountersAtLeast:  2,
						SourceLevelCountersLessThan: 4,
					}),
				},
				game.StaticAbility{
					Condition: opt.Val(game.Condition{
						SourceLevelCountersAtLeast: 4,
					}),
					ContinuousEffects: []game.ContinuousEffect{
						game.ContinuousEffect{
							Layer:          game.LayerPowerToughnessSet,
							AffectedSource: true,
							SetPower:       opt.Val(game.PT{Value: 4}),
							SetToughness:   opt.Val(game.PT{Value: 4}),
						},
					},
				},
				game.StaticAbility{
					KeywordAbilities: []game.KeywordAbility{
						game.SimpleKeyword{Kind: game.Flying},
					},
					Condition: opt.Val(game.Condition{
						SourceLevelCountersAtLeast: 4,
					}),
				},
				game.StaticAbility{
					Condition: opt.Val(game.Condition{
						SourceLevelCountersAtLeast: 4,
					}),
					ContinuousEffects: []game.ContinuousEffect{
						game.ContinuousEffect{
							Layer:          game.LayerPowerToughnessModify,
							Group:          game.ObjectControlledGroupExcluding(game.SourcePermanentReference(), game.Selection{SubtypesAny: []types.Sub{types.Sub("Merfolk")}}, game.SourcePermanentReference()),
							PowerDelta:     1,
							ToughnessDelta: 1,
						},
					},
				},
			},
			ActivatedAbilities: []game.ActivatedAbility{
				game.ActivatedAbility{
					Text:     "Level up {1} ({1}: Put a level counter on this. Level up only as a sorcery.)",
					ManaCost: opt.Val(cost.Mana{cost.O(1)}),
					Timing:   game.SorceryOnly,
					Content: game.Mode{
						Sequence: []game.Instruction{
							{
								Primitive: game.AddCounter{
									Amount:      game.Fixed(1),
									Object:      game.SourcePermanentReference(),
									CounterKind: counter.Level,
								},
							},
						},
					}.Ability(),
				},
			},
			OracleText: `
			Level up {1} ({1}: Put a level counter on this. Level up only as a sorcery.)
			LEVEL 2-3
			3/3
			Flying
			LEVEL 4+
			4/4
			Flying
			Other Merfolk creatures you control get +1/+1.
		`,
		},
	}
}
