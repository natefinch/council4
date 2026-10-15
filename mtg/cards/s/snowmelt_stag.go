package s

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// SnowmeltStag is the card definition for Snowmelt Stag.
//
// Type: Creature — Elemental Elk
// Cost: {3}{U}
//
// Oracle text:
//
//	Vigilance
//	During your turn, this creature has base power and toughness 5/2.
//	{5}{U}{U}: This creature can't be blocked this turn.
var SnowmeltStag = newSnowmeltStag

func newSnowmeltStag() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue),
		CardFace: game.CardFace{
			Name: "Snowmelt Stag",
			ManaCost: opt.Val(cost.Mana{
				cost.O(3),
				cost.U,
			}),
			Colors:    []color.Color{color.Blue},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Elemental, types.Elk},
			Power:     opt.Val(game.PT{Value: 2}),
			Toughness: opt.Val(game.PT{Value: 5}),
			StaticAbilities: []game.StaticAbility{
				game.VigilanceStaticBody,
				game.StaticAbility{
					Text: "During your turn, this creature has base power and toughness 5/2.",
					Condition: opt.Val(game.Condition{
						Text:                 "During your turn",
						SourceCounterKind:    counter.PlusOnePlusOne,
						SourceControllerTurn: true,
					}),
					ContinuousEffects: []game.ContinuousEffect{
						game.ContinuousEffect{
							Duration:       game.DurationPermanent,
							AffectedSource: true,
							Layer:          game.LayerPowerToughnessSet,
							SetPower: opt.Val(game.PT{
								Value: 5,
							}),
							SetToughness: opt.Val(game.PT{
								Value: 2,
							}),
						},
					},
				},
			},
			ActivatedAbilities: []game.ActivatedAbility{
				game.ActivatedAbility{
					Text: "{5}{U}{U}: This creature can't be blocked this turn.",
					ManaCost: opt.Val(cost.Mana{
						cost.O(5),
						cost.U,
						cost.U,
					}),
					ZoneOfFunction: zone.Battlefield,
					Content: game.AbilityContent{
						Modes: []game.Mode{
							game.Mode{
								Sequence: []game.Instruction{
									game.Instruction{
										Primitive: game.ApplyRule{
											Object: opt.Val(game.SourcePermanentReference()),
											RuleEffects: []game.RuleEffect{
												game.RuleEffect{
													Kind:     game.RuleEffectCantBeBlocked,
													Duration: game.DurationPermanent,
												},
											},
											Duration: game.DurationThisTurn,
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
			Vigilance
			During your turn, this creature has base power and toughness 5/2.
			{5}{U}{U}: This creature can't be blocked this turn.
		`,
		},
	}
}
