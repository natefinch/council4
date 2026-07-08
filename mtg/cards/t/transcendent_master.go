package t

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// TranscendentMaster is the card definition for Transcendent Master.
//
// Type: Creature — Human Cleric Avatar
// Cost: {1}{W}{W}
//
// Oracle text:
//
//	Level up {1} ({1}: Put a level counter on this. Level up only as a sorcery.)
//	LEVEL 6-11
//	6/6
//	Lifelink
//	LEVEL 12+
//	9/9
//	Lifelink, indestructible
var TranscendentMaster = newTranscendentMaster

func newTranscendentMaster() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White),
		CardFace: game.CardFace{
			Name: "Transcendent Master",
			ManaCost: opt.Val(cost.Mana{
				cost.O(1),
				cost.W,
				cost.W,
			}),
			Colors:    []color.Color{color.White},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Human, types.Cleric, types.Avatar},
			Power:     opt.Val(game.PT{Value: 3}),
			Toughness: opt.Val(game.PT{Value: 3}),
			StaticAbilities: []game.StaticAbility{
				game.StaticAbility{
					Condition: opt.Val(game.Condition{
						SourceLevelCountersAtLeast:  6,
						SourceLevelCountersLessThan: 12,
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
						game.SimpleKeyword{Kind: game.Lifelink},
					},
					Condition: opt.Val(game.Condition{
						SourceLevelCountersAtLeast:  6,
						SourceLevelCountersLessThan: 12,
					}),
				},
				game.StaticAbility{
					Condition: opt.Val(game.Condition{
						SourceLevelCountersAtLeast: 12,
					}),
					ContinuousEffects: []game.ContinuousEffect{
						game.ContinuousEffect{
							Layer:          game.LayerPowerToughnessSet,
							AffectedSource: true,
							SetPower:       opt.Val(game.PT{Value: 9}),
							SetToughness:   opt.Val(game.PT{Value: 9}),
						},
					},
				},
				game.StaticAbility{
					KeywordAbilities: []game.KeywordAbility{
						game.SimpleKeyword{Kind: game.Lifelink},
					},
					Condition: opt.Val(game.Condition{
						SourceLevelCountersAtLeast: 12,
					}),
				},
				game.StaticAbility{
					KeywordAbilities: []game.KeywordAbility{
						game.SimpleKeyword{Kind: game.Indestructible},
					},
					Condition: opt.Val(game.Condition{
						SourceLevelCountersAtLeast: 12,
					}),
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
			LEVEL 6-11
			6/6
			Lifelink
			LEVEL 12+
			9/9
			Lifelink, indestructible
		`,
		},
	}
}
