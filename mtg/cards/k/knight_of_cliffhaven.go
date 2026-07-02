package k

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// KnightOfCliffhaven is the card definition for Knight of Cliffhaven.
//
// Type: Creature — Kor Knight
// Cost: {1}{W}
//
// Oracle text:
//
//	Level up {3} ({3}: Put a level counter on this. Level up only as a sorcery.)
//	LEVEL 1-3
//	2/3
//	Flying
//	LEVEL 4+
//	4/4
//	Flying, vigilance
var KnightOfCliffhaven = newKnightOfCliffhaven()

func newKnightOfCliffhaven() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White),
		CardFace: game.CardFace{
			Name: "Knight of Cliffhaven",
			ManaCost: opt.Val(cost.Mana{
				cost.O(1),
				cost.W,
			}),
			Colors:    []color.Color{color.White},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Kor, types.Knight},
			Power:     opt.Val(game.PT{Value: 2}),
			Toughness: opt.Val(game.PT{Value: 2}),
			StaticAbilities: []game.StaticAbility{
				game.StaticAbility{
					Condition: opt.Val(game.Condition{
						SourceLevelCountersAtLeast:  1,
						SourceLevelCountersLessThan: 4,
					}),
					ContinuousEffects: []game.ContinuousEffect{
						game.ContinuousEffect{
							Layer:          game.LayerPowerToughnessSet,
							AffectedSource: true,
							SetPower:       opt.Val(game.PT{Value: 2}),
							SetToughness:   opt.Val(game.PT{Value: 3}),
						},
					},
				},
				game.StaticAbility{
					KeywordAbilities: []game.KeywordAbility{
						game.SimpleKeyword{Kind: game.Flying},
					},
					Condition: opt.Val(game.Condition{
						SourceLevelCountersAtLeast:  1,
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
					KeywordAbilities: []game.KeywordAbility{
						game.SimpleKeyword{Kind: game.Vigilance},
					},
					Condition: opt.Val(game.Condition{
						SourceLevelCountersAtLeast: 4,
					}),
				},
			},
			ActivatedAbilities: []game.ActivatedAbility{
				game.ActivatedAbility{
					Text:     "Level up {3} ({3}: Put a level counter on this. Level up only as a sorcery.)",
					ManaCost: opt.Val(cost.Mana{cost.O(3)}),
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
			Level up {3} ({3}: Put a level counter on this. Level up only as a sorcery.)
			LEVEL 1-3
			2/3
			Flying
			LEVEL 4+
			4/4
			Flying, vigilance
		`,
		},
	}
}
