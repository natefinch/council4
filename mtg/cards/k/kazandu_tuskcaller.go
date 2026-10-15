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

// KazanduTuskcaller is the card definition for Kazandu Tuskcaller.
//
// Type: Creature — Human Shaman
// Cost: {1}{G}
//
// Oracle text:
//
//	Level up {1}{G} ({1}{G}: Put a level counter on this. Level up only as a sorcery.)
//	LEVEL 2-5
//	1/1
//	{T}: Create a 3/3 green Elephant creature token.
//	LEVEL 6+
//	1/1
//	{T}: Create two 3/3 green Elephant creature tokens.
var KazanduTuskcaller = newKazanduTuskcaller

func newKazanduTuskcaller() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Green),
		CardFace: game.CardFace{
			Name: "Kazandu Tuskcaller",
			ManaCost: opt.Val(cost.Mana{
				cost.O(1),
				cost.G,
			}),
			Colors:    []color.Color{color.Green},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Human, types.Shaman},
			Power:     opt.Val(game.PT{Value: 1}),
			Toughness: opt.Val(game.PT{Value: 1}),
			StaticAbilities: []game.StaticAbility{
				game.StaticAbility{
					Text: "LEVEL 2-5\n1/1",
					Condition: opt.Val(game.Condition{
						SourceLevelCountersAtLeast:  2,
						SourceLevelCountersLessThan: 6,
						SourceCounterKind:           counter.PlusOnePlusOne,
					}),
					ContinuousEffects: []game.ContinuousEffect{
						game.ContinuousEffect{
							Duration:       game.DurationPermanent,
							AffectedSource: true,
							Layer:          game.LayerPowerToughnessSet,
							SetPower: opt.Val(game.PT{
								Value: 1,
							}),
							SetToughness: opt.Val(game.PT{
								Value: 1,
							}),
						},
					},
				},
				game.StaticAbility{
					Text: "LEVEL 6+\n1/1",
					Condition: opt.Val(game.Condition{
						SourceLevelCountersAtLeast: 6,
						SourceCounterKind:          counter.PlusOnePlusOne,
					}),
					ContinuousEffects: []game.ContinuousEffect{
						game.ContinuousEffect{
							Duration:       game.DurationPermanent,
							AffectedSource: true,
							Layer:          game.LayerPowerToughnessSet,
							SetPower: opt.Val(game.PT{
								Value: 1,
							}),
							SetToughness: opt.Val(game.PT{
								Value: 1,
							}),
						},
					},
				},
			},
			ActivatedAbilities: []game.ActivatedAbility{
				game.ActivatedAbility{
					Text: "Level up {1}{G} ({1}{G}: Put a level counter on this. Level up only as a sorcery.)",
					ManaCost: opt.Val(cost.Mana{
						cost.O(1),
						cost.G,
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
				game.ActivatedAbility{
					Text: "{T}: Create a 3/3 green Elephant creature token.",
					AdditionalCosts: []cost.Additional{
						cost.Additional{
							Kind:        cost.AdditionalTap,
							CounterKind: counter.PlusOnePlusOne,
						},
					},
					ZoneOfFunction: zone.Battlefield,
					ActivationCondition: opt.Val(game.Condition{
						SourceLevelCountersAtLeast:  2,
						SourceLevelCountersLessThan: 6,
						SourceCounterKind:           counter.PlusOnePlusOne,
					}),
					Content: game.AbilityContent{
						Modes: []game.Mode{
							game.Mode{
								Sequence: []game.Instruction{
									game.Instruction{
										Primitive: game.CreateToken{
											Amount: game.Fixed(1),
											Source: game.TokenDef(kazanduTuskcallerToken),
										},
									},
								},
							},
						},
						MinModes: 1,
						MaxModes: 1,
					},
				},
				game.ActivatedAbility{
					Text: "{T}: Create two 3/3 green Elephant creature tokens.",
					AdditionalCosts: []cost.Additional{
						cost.Additional{
							Kind:        cost.AdditionalTap,
							CounterKind: counter.PlusOnePlusOne,
						},
					},
					ZoneOfFunction: zone.Battlefield,
					ActivationCondition: opt.Val(game.Condition{
						SourceLevelCountersAtLeast: 6,
						SourceCounterKind:          counter.PlusOnePlusOne,
					}),
					Content: game.AbilityContent{
						Modes: []game.Mode{
							game.Mode{
								Sequence: []game.Instruction{
									game.Instruction{
										Primitive: game.CreateToken{
											Amount: game.Fixed(2),
											Source: game.TokenDef(kazanduTuskcallerToken),
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
			Level up {1}{G} ({1}{G}: Put a level counter on this. Level up only as a sorcery.)
			LEVEL 2-5
			1/1
			{T}: Create a 3/3 green Elephant creature token.
			LEVEL 6+
			1/1
			{T}: Create two 3/3 green Elephant creature tokens.
		`,
		},
	}
}

var kazanduTuskcallerToken = newKazanduTuskcallerToken()

func newKazanduTuskcallerToken() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name:      "Elephant",
			Colors:    []color.Color{color.Green},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Elephant},
			Power:     opt.Val(game.PT{Value: 3}),
			Toughness: opt.Val(game.PT{Value: 3}),
		},
	}
}
