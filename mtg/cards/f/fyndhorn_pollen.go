package f

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// FyndhornPollen is the card definition for Fyndhorn Pollen.
//
// Type: Enchantment
// Cost: {2}{G}
//
// Oracle text:
//
//	Cumulative upkeep {1} (At the beginning of your upkeep, put an age counter on this permanent, then sacrifice it unless you pay its upkeep cost for each age counter on it.)
//	All creatures get -1/-0.
//	{1}{G}: All creatures get -1/-0 until end of turn.
var FyndhornPollen = newFyndhornPollen

func newFyndhornPollen() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Green),
		CardFace: game.CardFace{
			Name: "Fyndhorn Pollen",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
				cost.G,
			}),
			Colors: []color.Color{color.Green},
			Types:  []types.Card{types.Enchantment},
			StaticAbilities: []game.StaticAbility{
				game.StaticAbility{
					Text: "All creatures get -1/-0.",
					ContinuousEffects: []game.ContinuousEffect{
						game.ContinuousEffect{
							Duration:   game.DurationPermanent,
							Group:      game.BattlefieldGroup(game.Selection{RequiredTypes: []types.Card{types.Creature}}),
							Layer:      game.LayerPowerToughnessModify,
							PowerDelta: -1,
						},
					},
				},
			},
			ActivatedAbilities: []game.ActivatedAbility{
				game.ActivatedAbility{
					Text: "{1}{G}: All creatures get -1/-0 until end of turn.",
					ManaCost: opt.Val(cost.Mana{
						cost.O(1),
						cost.G,
					}),
					ZoneOfFunction: zone.Battlefield,
					Content: game.AbilityContent{
						Modes: []game.Mode{
							game.Mode{
								Sequence: []game.Instruction{
									game.Instruction{
										Primitive: game.ApplyContinuous{
											ContinuousEffects: []game.ContinuousEffect{
												game.ContinuousEffect{
													Duration:   game.DurationPermanent,
													Group:      game.BattlefieldGroup(game.Selection{RequiredTypes: []types.Card{types.Creature}}),
													Layer:      game.LayerPowerToughnessModify,
													PowerDelta: -1,
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
			TriggeredAbilities: []game.TriggeredAbility{
				game.CumulativeUpkeepTriggeredAbility(cost.Mana{cost.O(1)}),
			},
			OracleText: `
			Cumulative upkeep {1} (At the beginning of your upkeep, put an age counter on this permanent, then sacrifice it unless you pay its upkeep cost for each age counter on it.)
			All creatures get -1/-0.
			{1}{G}: All creatures get -1/-0 until end of turn.
		`,
		},
	}
}
