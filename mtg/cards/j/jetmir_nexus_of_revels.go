package j

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// JetmirNexusOfRevels is the card definition for Jetmir, Nexus of Revels.
//
// Type: Legendary Creature — Cat Demon
// Cost: {1}{R}{G}{W}
//
// Oracle text:
//
//	Creatures you control get +1/+0 and have vigilance as long as you control three or more creatures.
//	Creatures you control also get +1/+0 and have trample as long as you control six or more creatures.
//	Creatures you control also get +1/+0 and have double strike as long as you control nine or more creatures.
var JetmirNexusOfRevels = newJetmirNexusOfRevels

func newJetmirNexusOfRevels() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White, color.Red, color.Green),
		CardFace: game.CardFace{
			Name: "Jetmir, Nexus of Revels",
			ManaCost: opt.Val(cost.Mana{
				cost.O(1),
				cost.R,
				cost.G,
				cost.W,
			}),
			Colors:     []color.Color{color.Green, color.Red, color.White},
			Supertypes: []types.Super{types.Legendary},
			Types:      []types.Card{types.Creature},
			Subtypes:   []types.Sub{types.Cat, types.Demon},
			Power:      opt.Val(game.PT{Value: 5}),
			Toughness:  opt.Val(game.PT{Value: 4}),
			StaticAbilities: []game.StaticAbility{
				game.StaticAbility{
					Text: "Creatures you control get +1/+0 and have vigilance as long as you control three or more creatures.",
					Condition: opt.Val(game.Condition{
						Text: "as long as you control three or more creatures",
						ControlsMatching: opt.Val(game.SelectionCount{
							Selection: game.Selection{RequiredTypes: []types.Card{types.Creature}},
							MinCount:  3,
						}),
						SourceCounterKind: counter.PlusOnePlusOne,
					}),
					ContinuousEffects: []game.ContinuousEffect{
						game.ContinuousEffect{
							Duration:   game.DurationPermanent,
							Group:      game.ObjectControlledGroup(game.SourcePermanentReference(), game.Selection{RequiredTypes: []types.Card{types.Creature}}),
							Layer:      game.LayerPowerToughnessModify,
							PowerDelta: 1,
						},
						game.ContinuousEffect{
							Duration:    game.DurationPermanent,
							Group:       game.ObjectControlledGroup(game.SourcePermanentReference(), game.Selection{RequiredTypes: []types.Card{types.Creature}}),
							Layer:       game.LayerAbility,
							AddKeywords: []game.Keyword{game.Vigilance},
						},
					},
				},
				game.StaticAbility{
					Text: "Creatures you control also get +1/+0 and have trample as long as you control six or more creatures.",
					Condition: opt.Val(game.Condition{
						Text: "as long as you control six or more creatures",
						ControlsMatching: opt.Val(game.SelectionCount{
							Selection: game.Selection{RequiredTypes: []types.Card{types.Creature}},
							MinCount:  6,
						}),
						SourceCounterKind: counter.PlusOnePlusOne,
					}),
					ContinuousEffects: []game.ContinuousEffect{
						game.ContinuousEffect{
							Duration:   game.DurationPermanent,
							Group:      game.ObjectControlledGroup(game.SourcePermanentReference(), game.Selection{RequiredTypes: []types.Card{types.Creature}}),
							Layer:      game.LayerPowerToughnessModify,
							PowerDelta: 1,
						},
						game.ContinuousEffect{
							Duration:    game.DurationPermanent,
							Group:       game.ObjectControlledGroup(game.SourcePermanentReference(), game.Selection{RequiredTypes: []types.Card{types.Creature}}),
							Layer:       game.LayerAbility,
							AddKeywords: []game.Keyword{game.Trample},
						},
					},
				},
				game.StaticAbility{
					Text: "Creatures you control also get +1/+0 and have double strike as long as you control nine or more creatures.",
					Condition: opt.Val(game.Condition{
						Text: "as long as you control nine or more creatures",
						ControlsMatching: opt.Val(game.SelectionCount{
							Selection: game.Selection{RequiredTypes: []types.Card{types.Creature}},
							MinCount:  9,
						}),
						SourceCounterKind: counter.PlusOnePlusOne,
					}),
					ContinuousEffects: []game.ContinuousEffect{
						game.ContinuousEffect{
							Duration:   game.DurationPermanent,
							Group:      game.ObjectControlledGroup(game.SourcePermanentReference(), game.Selection{RequiredTypes: []types.Card{types.Creature}}),
							Layer:      game.LayerPowerToughnessModify,
							PowerDelta: 1,
						},
						game.ContinuousEffect{
							Duration:    game.DurationPermanent,
							Group:       game.ObjectControlledGroup(game.SourcePermanentReference(), game.Selection{RequiredTypes: []types.Card{types.Creature}}),
							Layer:       game.LayerAbility,
							AddKeywords: []game.Keyword{game.DoubleStrike},
						},
					},
				},
			},
			OracleText: `
			Creatures you control get +1/+0 and have vigilance as long as you control three or more creatures.
			Creatures you control also get +1/+0 and have trample as long as you control six or more creatures.
			Creatures you control also get +1/+0 and have double strike as long as you control nine or more creatures.
		`,
		},
	}
}
