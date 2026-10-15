package m

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/compare"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// MindDrillAssailant is the card definition for Mind Drill Assailant.
//
// Type: Creature — Rat Warlock
// Cost: {2}{U/B}{U/B}
//
// Oracle text:
//
//	Threshold — As long as there are seven or more cards in your graveyard, this creature gets +3/+0.
//	{2}{U/B}: Surveil 1. (Look at the top card of your library. You may put it into your graveyard.)
var MindDrillAssailant = newMindDrillAssailant

func newMindDrillAssailant() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue, color.Black),
		CardFace: game.CardFace{
			Name: "Mind Drill Assailant",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
				cost.HybridMana(mana.U, mana.B),
				cost.HybridMana(mana.U, mana.B),
			}),
			Colors:    []color.Color{color.Black, color.Blue},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Rat, types.Warlock},
			Power:     opt.Val(game.PT{Value: 2}),
			Toughness: opt.Val(game.PT{Value: 5}),
			StaticAbilities: []game.StaticAbility{
				game.StaticAbility{
					Text: "Threshold — As long as there are seven or more cards in your graveyard, this creature gets +3/+0.",
					Condition: opt.Val(game.Condition{
						Text: "As long as there are seven or more cards in your graveyard",
						Aggregates: []game.AggregateComparison{
							game.AggregateComparison{
								Aggregate: game.AggregateControllerGraveyardCardCount,
								Op:        compare.GreaterOrEqual,
								Value:     7,
							},
						},
						SourceCounterKind: counter.PlusOnePlusOne,
					}),
					ContinuousEffects: []game.ContinuousEffect{
						game.ContinuousEffect{
							Duration:       game.DurationPermanent,
							AffectedSource: true,
							Layer:          game.LayerPowerToughnessModify,
							PowerDelta:     3,
						},
					},
				},
			},
			ActivatedAbilities: []game.ActivatedAbility{
				game.ActivatedAbility{
					Text: "{2}{U/B}: Surveil 1. (Look at the top card of your library. You may put it into your graveyard.)",
					ManaCost: opt.Val(cost.Mana{
						cost.O(2),
						cost.HybridMana(mana.U, mana.B),
					}),
					ZoneOfFunction: zone.Battlefield,
					Content: game.AbilityContent{
						Modes: []game.Mode{
							game.Mode{
								Sequence: []game.Instruction{
									game.Instruction{
										Primitive: game.Surveil{
											Amount: game.Fixed(1),
											Player: game.ControllerReference(),
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
			Threshold — As long as there are seven or more cards in your graveyard, this creature gets +3/+0.
			{2}{U/B}: Surveil 1. (Look at the top card of your library. You may put it into your graveyard.)
		`,
		},
	}
}
