package n

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// NirkanaCutthroat is the card definition for Nirkana Cutthroat.
//
// Type: Creature — Vampire Warrior
// Cost: {2}{B}
//
// Oracle text:
//
//	Level up {2}{B} ({2}{B}: Put a level counter on this. Level up only as a sorcery.)
//	LEVEL 1-2
//	4/3
//	Deathtouch
//	LEVEL 3+
//	5/4
//	First strike, deathtouch
var NirkanaCutthroat = newNirkanaCutthroat()

func newNirkanaCutthroat() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Black),
		CardFace: game.CardFace{
			Name: "Nirkana Cutthroat",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
				cost.B,
			}),
			Colors:    []color.Color{color.Black},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Vampire, types.Warrior},
			Power:     opt.Val(game.PT{Value: 3}),
			Toughness: opt.Val(game.PT{Value: 2}),
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
							SetPower:       opt.Val(game.PT{Value: 4}),
							SetToughness:   opt.Val(game.PT{Value: 3}),
						},
					},
				},
				game.StaticAbility{
					KeywordAbilities: []game.KeywordAbility{
						game.SimpleKeyword{Kind: game.Deathtouch},
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
							SetPower:       opt.Val(game.PT{Value: 5}),
							SetToughness:   opt.Val(game.PT{Value: 4}),
						},
					},
				},
				game.StaticAbility{
					KeywordAbilities: []game.KeywordAbility{
						game.SimpleKeyword{Kind: game.FirstStrike},
					},
					Condition: opt.Val(game.Condition{
						SourceLevelCountersAtLeast: 3,
					}),
				},
				game.StaticAbility{
					KeywordAbilities: []game.KeywordAbility{
						game.SimpleKeyword{Kind: game.Deathtouch},
					},
					Condition: opt.Val(game.Condition{
						SourceLevelCountersAtLeast: 3,
					}),
				},
			},
			ActivatedAbilities: []game.ActivatedAbility{
				game.ActivatedAbility{
					Text:     "Level up {2}{B} ({2}{B}: Put a level counter on this. Level up only as a sorcery.)",
					ManaCost: opt.Val(cost.Mana{cost.O(2), cost.B}),
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
			Level up {2}{B} ({2}{B}: Put a level counter on this. Level up only as a sorcery.)
			LEVEL 1-2
			4/3
			Deathtouch
			LEVEL 3+
			5/4
			First strike, deathtouch
		`,
		},
	}
}
