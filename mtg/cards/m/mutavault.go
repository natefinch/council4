package m

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// Mutavault is the card definition for Mutavault.
//
// Type: Land
//
// Oracle text:
//
//	{T}: Add {C}.
//	{1}: This land becomes a 2/2 creature with all creature types until end of turn. It's still a land.
var Mutavault = newMutavault

func newMutavault() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name:  "Mutavault",
			Types: []types.Card{types.Land},
			ActivatedAbilities: []game.ActivatedAbility{
				game.ActivatedAbility{
					Text: "{1}: This land becomes a 2/2 creature with all creature types until end of turn. It's still a land.",
					ManaCost: opt.Val(cost.Mana{
						cost.O(1),
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
													Duration:             game.DurationPermanent,
													Layer:                game.LayerType,
													AddTypes:             []types.Card{types.Creature},
													AddEveryCreatureType: true,
												},
												game.ContinuousEffect{
													Duration: game.DurationPermanent,
													Layer:    game.LayerPowerToughnessSet,
													SetPower: opt.Val(game.PT{
														Value: 2,
													}),
													SetToughness: opt.Val(game.PT{
														Value: 2,
													}),
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
			ManaAbilities: []game.ManaAbility{
				game.TapManaAbility(mana.C),
			},
			OracleText: `
			{T}: Add {C}.
			{1}: This land becomes a 2/2 creature with all creature types until end of turn. It's still a land.
		`,
		},
	}
}
