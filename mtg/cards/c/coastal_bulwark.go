package c

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// CoastalBulwark is the card definition for Coastal Bulwark.
//
// Type: Artifact Creature — Wall
// Cost: {2}
//
// Oracle text:
//
//	Defender
//	This creature gets +2/+0 as long as you control an Island.
//	{2}, {T}: Surveil 1. (Look at the top card of your library. You may put that card into your graveyard.)
var CoastalBulwark = newCoastalBulwark

func newCoastalBulwark() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name: "Coastal Bulwark",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
			}),
			Types:     []types.Card{types.Artifact, types.Creature},
			Subtypes:  []types.Sub{types.Wall},
			Power:     opt.Val(game.PT{Value: 1}),
			Toughness: opt.Val(game.PT{Value: 3}),
			StaticAbilities: []game.StaticAbility{
				game.DefenderStaticBody,
				game.StaticAbility{
					Text: "This creature gets +2/+0 as long as you control an Island.",
					Condition: opt.Val(game.Condition{
						Text: "as long as you control an Island",
						ControlsMatching: opt.Val(game.SelectionCount{
							Selection: game.Selection{SubtypesAny: []types.Sub{types.Island}},
						}),
						SourceCounterKind: counter.PlusOnePlusOne,
					}),
					ContinuousEffects: []game.ContinuousEffect{
						game.ContinuousEffect{
							Duration:       game.DurationPermanent,
							AffectedSource: true,
							Layer:          game.LayerPowerToughnessModify,
							PowerDelta:     2,
						},
					},
				},
			},
			ActivatedAbilities: []game.ActivatedAbility{
				game.ActivatedAbility{
					Text: "{2}, {T}: Surveil 1. (Look at the top card of your library. You may put that card into your graveyard.)",
					ManaCost: opt.Val(cost.Mana{
						cost.O(2),
					}),
					AdditionalCosts: []cost.Additional{
						cost.Additional{
							Kind:        cost.AdditionalTap,
							CounterKind: counter.PlusOnePlusOne,
						},
					},
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
			Defender
			This creature gets +2/+0 as long as you control an Island.
			{2}, {T}: Surveil 1. (Look at the top card of your library. You may put that card into your graveyard.)
		`,
		},
	}
}
