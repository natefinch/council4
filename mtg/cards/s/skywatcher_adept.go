package s

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// SkywatcherAdept is the card definition for Skywatcher Adept.
//
// Type: Creature — Merfolk Wizard
// Cost: {U}
//
// Oracle text:
//
//	Level up {3} ({3}: Put a level counter on this. Level up only as a sorcery.)
//	LEVEL 1-2
//	2/2
//	Flying
//	LEVEL 3+
//	4/2
//	Flying
var SkywatcherAdept = newSkywatcherAdept

func newSkywatcherAdept() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue),
		CardFace: game.CardFace{
			Name: "Skywatcher Adept",
			ManaCost: opt.Val(cost.Mana{
				cost.U,
			}),
			Colors:    []color.Color{color.Blue},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Merfolk, types.Wizard},
			Power:     opt.Val(game.PT{Value: 1}),
			Toughness: opt.Val(game.PT{Value: 1}),
			StaticAbilities: []game.StaticAbility{
				game.StaticAbility{
					Condition: opt.Val(game.Condition{
						SourceLevelCountersAtLeast:  1,
						SourceLevelCountersLessThan: 3,
					}),
					ContinuousEffects: []game.ContinuousEffect{
						game.ContinuousEffect{
							Layer:          game.LayerPowerToughnessSet,
							AffectedSource: true,
							SetPower:       opt.Val(game.PT{Value: 2}),
							SetToughness:   opt.Val(game.PT{Value: 2}),
						},
					},
				},
				game.StaticAbility{
					KeywordAbilities: []game.KeywordAbility{
						game.SimpleKeyword{Kind: game.Flying},
					},
					Condition: opt.Val(game.Condition{
						SourceLevelCountersAtLeast:  1,
						SourceLevelCountersLessThan: 3,
					}),
				},
				game.StaticAbility{
					Condition: opt.Val(game.Condition{
						SourceLevelCountersAtLeast: 3,
					}),
					ContinuousEffects: []game.ContinuousEffect{
						game.ContinuousEffect{
							Layer:          game.LayerPowerToughnessSet,
							AffectedSource: true,
							SetPower:       opt.Val(game.PT{Value: 4}),
							SetToughness:   opt.Val(game.PT{Value: 2}),
						},
					},
				},
				game.StaticAbility{
					KeywordAbilities: []game.KeywordAbility{
						game.SimpleKeyword{Kind: game.Flying},
					},
					Condition: opt.Val(game.Condition{
						SourceLevelCountersAtLeast: 3,
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
			LEVEL 1-2
			2/2
			Flying
			LEVEL 3+
			4/2
			Flying
		`,
		},
	}
}
