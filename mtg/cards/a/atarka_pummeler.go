package a

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/compare"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// AtarkaPummeler is the card definition for Atarka Pummeler.
//
// Type: Creature — Ogre Warrior
// Cost: {4}{R}
//
// Oracle text:
//
//	Formidable — {3}{R}{R}: Creatures you control gain menace until end of turn. Activate only if creatures you control have total power 8 or greater. (They can't be blocked except by two or more creatures.)
var AtarkaPummeler = newAtarkaPummeler

func newAtarkaPummeler() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Red),
		CardFace: game.CardFace{
			Name: "Atarka Pummeler",
			ManaCost: opt.Val(cost.Mana{
				cost.O(4),
				cost.R,
			}),
			Colors:    []color.Color{color.Red},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Ogre, types.Warrior},
			Power:     opt.Val(game.PT{Value: 4}),
			Toughness: opt.Val(game.PT{Value: 5}),
			ActivatedAbilities: []game.ActivatedAbility{
				game.ActivatedAbility{
					Text: "Formidable — {3}{R}{R}: Creatures you control gain menace until end of turn. Activate only if creatures you control have total power 8 or greater. (They can't be blocked except by two or more creatures.)",
					ManaCost: opt.Val(cost.Mana{
						cost.O(3),
						cost.R,
						cost.R,
					}),
					ZoneOfFunction: zone.Battlefield,
					ActivationCondition: opt.Val(game.Condition{
						Text: "only if creatures you control have total power 8 or greater",
						ControlsMatching: opt.Val(game.SelectionCount{
							Selection: game.Selection{RequiredTypes: []types.Card{types.Creature}},
							TotalPower: opt.Val(compare.Int{
								Op:    compare.GreaterOrEqual,
								Value: 8,
							}),
						}),
						SourceCounterKind: counter.PlusOnePlusOne,
					}),
					Content: game.AbilityContent{
						Modes: []game.Mode{
							game.Mode{
								Sequence: []game.Instruction{
									game.Instruction{
										Primitive: game.ApplyContinuous{
											ContinuousEffects: []game.ContinuousEffect{
												game.ContinuousEffect{
													Duration:    game.DurationPermanent,
													Group:       game.BattlefieldGroup(game.Selection{RequiredTypes: []types.Card{types.Creature}, Controller: game.ControllerYou}),
													Layer:       game.LayerAbility,
													AddKeywords: []game.Keyword{game.Menace},
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
			Formidable — {3}{R}{R}: Creatures you control gain menace until end of turn. Activate only if creatures you control have total power 8 or greater. (They can't be blocked except by two or more creatures.)
		`,
		},
	}
}
