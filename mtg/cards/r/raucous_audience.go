package r

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/compare"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// RaucousAudience is the card definition for Raucous Audience.
//
// Type: Creature — Human Citizen
// Cost: {1}{G}
//
// Oracle text:
//
//	{T}: Add {G}. If you control a creature with power 4 or greater, add {G}{G} instead.
var RaucousAudience = newRaucousAudience

func newRaucousAudience() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Green),
		CardFace: game.CardFace{
			Name: "Raucous Audience",
			ManaCost: opt.Val(cost.Mana{
				cost.O(1),
				cost.G,
			}),
			Colors:    []color.Color{color.Green},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Human, types.Citizen},
			Power:     opt.Val(game.PT{Value: 2}),
			Toughness: opt.Val(game.PT{Value: 1}),
			ManaAbilities: []game.ManaAbility{
				game.ManaAbility{
					Text: "{T}: Add {G}. If you control a creature with power 4 or greater, add {G}{G} instead.",
					AdditionalCosts: []cost.Additional{
						cost.Additional{
							Kind:        cost.AdditionalTap,
							CounterKind: counter.PlusOnePlusOne,
						},
					},
					Content: game.AbilityContent{
						Modes: []game.Mode{
							game.Mode{
								Sequence: []game.Instruction{
									game.Instruction{
										Primitive: game.AddMana{
											Amount:    game.Fixed(1),
											ManaColor: mana.G,
										},
										Condition: opt.Val(game.EffectCondition{
											Condition: opt.Val(game.Condition{
												Text:   "If you control a creature with power 4 or greater",
												Negate: true,
												ControlsMatching: opt.Val(game.SelectionCount{
													Selection: game.Selection{RequiredTypes: []types.Card{types.Creature}, Power: opt.Val(compare.Int{
														Op:    compare.GreaterOrEqual,
														Value: 4,
													})},
												}),
												SourceCounterKind: counter.PlusOnePlusOne,
											}),
										}),
									},
									game.Instruction{
										Primitive: game.AddMana{
											Amount:    game.Fixed(1),
											ManaColor: mana.G,
										},
										Condition: opt.Val(game.EffectCondition{
											Condition: opt.Val(game.Condition{
												Text: "If you control a creature with power 4 or greater",
												ControlsMatching: opt.Val(game.SelectionCount{
													Selection: game.Selection{RequiredTypes: []types.Card{types.Creature}, Power: opt.Val(compare.Int{
														Op:    compare.GreaterOrEqual,
														Value: 4,
													})},
												}),
												SourceCounterKind: counter.PlusOnePlusOne,
											}),
										}),
									},
									game.Instruction{
										Primitive: game.AddMana{
											Amount:    game.Fixed(1),
											ManaColor: mana.G,
										},
										Condition: opt.Val(game.EffectCondition{
											Condition: opt.Val(game.Condition{
												Text: "If you control a creature with power 4 or greater",
												ControlsMatching: opt.Val(game.SelectionCount{
													Selection: game.Selection{RequiredTypes: []types.Card{types.Creature}, Power: opt.Val(compare.Int{
														Op:    compare.GreaterOrEqual,
														Value: 4,
													})},
												}),
												SourceCounterKind: counter.PlusOnePlusOne,
											}),
										}),
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
			{T}: Add {G}. If you control a creature with power 4 or greater, add {G}{G} instead.
		`,
		},
	}
}
