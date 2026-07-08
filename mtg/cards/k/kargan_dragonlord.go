package k

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// KarganDragonlord is the card definition for Kargan Dragonlord.
//
// Type: Creature — Human Warrior
// Cost: {R}{R}
//
// Oracle text:
//
//	Level up {R} ({R}: Put a level counter on this. Level up only as a sorcery.)
//	LEVEL 4-7
//	4/4
//	Flying
//	LEVEL 8+
//	8/8
//	Flying, trample
//	{R}: This creature gets +1/+0 until end of turn.
var KarganDragonlord = newKarganDragonlord

func newKarganDragonlord() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Red),
		CardFace: game.CardFace{
			Name: "Kargan Dragonlord",
			ManaCost: opt.Val(cost.Mana{
				cost.R,
				cost.R,
			}),
			Colors:    []color.Color{color.Red},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Human, types.Warrior},
			Power:     opt.Val(game.PT{Value: 2}),
			Toughness: opt.Val(game.PT{Value: 2}),
			StaticAbilities: []game.StaticAbility{
				game.StaticAbility{
					Condition: opt.Val(game.Condition{
						SourceLevelCountersAtLeast:  4,
						SourceLevelCountersLessThan: 8,
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
						SourceLevelCountersAtLeast:  4,
						SourceLevelCountersLessThan: 8,
					}),
				},
				game.StaticAbility{
					Condition: opt.Val(game.Condition{
						SourceLevelCountersAtLeast: 8,
					}),
					ContinuousEffects: []game.ContinuousEffect{
						game.ContinuousEffect{
							Layer:          game.LayerPowerToughnessSet,
							AffectedSource: true,
							SetPower:       opt.Val(game.PT{Value: 8}),
							SetToughness:   opt.Val(game.PT{Value: 8}),
						},
					},
				},
				game.StaticAbility{
					KeywordAbilities: []game.KeywordAbility{
						game.SimpleKeyword{Kind: game.Flying},
					},
					Condition: opt.Val(game.Condition{
						SourceLevelCountersAtLeast: 8,
					}),
				},
				game.StaticAbility{
					KeywordAbilities: []game.KeywordAbility{
						game.SimpleKeyword{Kind: game.Trample},
					},
					Condition: opt.Val(game.Condition{
						SourceLevelCountersAtLeast: 8,
					}),
				},
			},
			ActivatedAbilities: []game.ActivatedAbility{
				game.ActivatedAbility{
					Text:     "Level up {R} ({R}: Put a level counter on this. Level up only as a sorcery.)",
					ManaCost: opt.Val(cost.Mana{cost.R}),
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
				game.ActivatedAbility{
					Text:           "{R}: This creature gets +1/+0 until end of turn.",
					ManaCost:       opt.Val(cost.Mana{cost.R}),
					ZoneOfFunction: zone.Battlefield,
					ActivationCondition: opt.Val(game.Condition{
						SourceLevelCountersAtLeast: 8,
					}),
					Content: game.Mode{
						Sequence: []game.Instruction{
							{
								Primitive: game.ModifyPT{
									Object:         game.SourcePermanentReference(),
									PowerDelta:     game.Fixed(1),
									ToughnessDelta: game.Fixed(0),
									Duration:       game.DurationUntilEndOfTurn,
								},
							},
						},
					}.Ability(),
				},
			},
			OracleText: `
			Level up {R} ({R}: Put a level counter on this. Level up only as a sorcery.)
			LEVEL 4-7
			4/4
			Flying
			LEVEL 8+
			8/8
			Flying, trample
			{R}: This creature gets +1/+0 until end of turn.
		`,
		},
	}
}
