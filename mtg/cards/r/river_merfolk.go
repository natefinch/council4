package r

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// RiverMerfolk is the card definition for River Merfolk.
//
// Type: Creature — Merfolk
// Cost: {U}{U}
//
// Oracle text:
//
//	{U}: This creature gains mountainwalk until end of turn. (It can't be blocked as long as defending player controls a Mountain.)
var RiverMerfolk = newRiverMerfolk

func newRiverMerfolk() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue),
		CardFace: game.CardFace{
			Name: "River Merfolk",
			ManaCost: opt.Val(cost.Mana{
				cost.U,
				cost.U,
			}),
			Colors:    []color.Color{color.Blue},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Merfolk},
			Power:     opt.Val(game.PT{Value: 2}),
			Toughness: opt.Val(game.PT{Value: 1}),
			ActivatedAbilities: []game.ActivatedAbility{
				game.ActivatedAbility{
					Text: "{U}: This creature gains mountainwalk until end of turn. (It can't be blocked as long as defending player controls a Mountain.)",
					ManaCost: opt.Val(cost.Mana{
						cost.U,
					}),
					ZoneOfFunction: zone.Battlefield,
					Content: game.AbilityContent{
						Modes: []game.Mode{
							game.Mode{
								Sequence: []game.Instruction{
									game.Instruction{
										Primitive: game.ApplyContinuous{
											Object: opt.Val(game.SourcePermanentReference()),
											ContinuousEffects: []game.ContinuousEffect{
												game.ContinuousEffect{
													Duration: game.DurationPermanent,
													Layer:    game.LayerAbility,
													AddAbilities: []game.Ability{
														&game.StaticAbility{
															Text: "Mountainwalk",
															KeywordAbilities: []game.KeywordAbility{
																game.LandwalkKeyword{
																	Subtype: types.Mountain,
																},
															},
														},
													},
												},
											},
											Duration: game.DurationUntilEndOfTurn,
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
			{U}: This creature gains mountainwalk until end of turn. (It can't be blocked as long as defending player controls a Mountain.)
		`,
		},
	}
}
