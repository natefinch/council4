package h

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// HalimarWavewatch is the card definition for Halimar Wavewatch.
//
// Type: Creature — Merfolk Soldier
// Cost: {1}{U}
//
// Oracle text:
//
//	Level up {2} ({2}: Put a level counter on this. Level up only as a sorcery.)
//	LEVEL 1-4
//	0/6
//	LEVEL 5+
//	6/6
//	Islandwalk (This creature can't be blocked as long as defending player controls an Island.)
var HalimarWavewatch = newHalimarWavewatch

func newHalimarWavewatch() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue),
		CardFace: game.CardFace{
			Name: "Halimar Wavewatch",
			ManaCost: opt.Val(cost.Mana{
				cost.O(1),
				cost.U,
			}),
			Colors:    []color.Color{color.Blue},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Merfolk, types.Soldier},
			Power:     opt.Val(game.PT{Value: 0}),
			Toughness: opt.Val(game.PT{Value: 3}),
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
							SetPower:       opt.Val(game.PT{Value: 0}),
							SetToughness:   opt.Val(game.PT{Value: 6}),
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
							SetPower:       opt.Val(game.PT{Value: 6}),
							SetToughness:   opt.Val(game.PT{Value: 6}),
						},
					},
				},
				game.StaticAbility{
					KeywordAbilities: []game.KeywordAbility{
						game.LandwalkKeyword{Subtype: types.Island},
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
			0/6
			LEVEL 5+
			6/6
			Islandwalk (This creature can't be blocked as long as defending player controls an Island.)
		`,
		},
	}
}
