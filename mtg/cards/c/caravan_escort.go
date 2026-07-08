package c

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// CaravanEscort is the card definition for Caravan Escort.
//
// Type: Creature — Human Knight
// Cost: {W}
//
// Oracle text:
//
//	Level up {2} ({2}: Put a level counter on this. Level up only as a sorcery.)
//	LEVEL 1-4
//	2/2
//	LEVEL 5+
//	5/5
//	First strike
var CaravanEscort = newCaravanEscort

func newCaravanEscort() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White),
		CardFace: game.CardFace{
			Name: "Caravan Escort",
			ManaCost: opt.Val(cost.Mana{
				cost.W,
			}),
			Colors:    []color.Color{color.White},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Human, types.Knight},
			Power:     opt.Val(game.PT{Value: 1}),
			Toughness: opt.Val(game.PT{Value: 1}),
			StaticAbilities: []game.StaticAbility{
				game.StaticAbility{
					Condition: opt.Val(game.Condition{
						SourceLevelCountersAtLeast:  1,
						SourceLevelCountersLessThan: 5,
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
					Condition: opt.Val(game.Condition{
						SourceLevelCountersAtLeast: 5,
					}),
					ContinuousEffects: []game.ContinuousEffect{
						game.ContinuousEffect{
							Layer:          game.LayerPowerToughnessSet,
							AffectedSource: true,
							SetPower:       opt.Val(game.PT{Value: 5}),
							SetToughness:   opt.Val(game.PT{Value: 5}),
						},
					},
				},
				game.StaticAbility{
					KeywordAbilities: []game.KeywordAbility{
						game.SimpleKeyword{Kind: game.FirstStrike},
					},
					Condition: opt.Val(game.Condition{
						SourceLevelCountersAtLeast: 5,
					}),
				},
			},
			ActivatedAbilities: []game.ActivatedAbility{
				game.ActivatedAbility{
					Text:     "Level up {2} ({2}: Put a level counter on this. Level up only as a sorcery.)",
					ManaCost: opt.Val(cost.Mana{cost.O(2)}),
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
			Level up {2} ({2}: Put a level counter on this. Level up only as a sorcery.)
			LEVEL 1-4
			2/2
			LEVEL 5+
			5/5
			First strike
		`,
		},
	}
}
