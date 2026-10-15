package d

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// DreadStatuary is the card definition for Dread Statuary.
//
// Type: Land
//
// Oracle text:
//
//	{T}: Add {C}.
//	{4}: This land becomes a 4/2 Golem artifact creature until end of turn. It's still a land.
var DreadStatuary = newDreadStatuary

func newDreadStatuary() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name:  "Dread Statuary",
			Types: []types.Card{types.Land},
			ActivatedAbilities: []game.ActivatedAbility{
				game.ActivatedAbility{
					Text: "{4}: This land becomes a 4/2 Golem artifact creature until end of turn. It's still a land.",
					ManaCost: opt.Val(cost.Mana{
						cost.O(4),
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
													Duration:    game.DurationPermanent,
													Layer:       game.LayerType,
													AddTypes:    []types.Card{types.Creature, types.Artifact},
													AddSubtypes: []types.Sub{types.Golem},
												},
												game.ContinuousEffect{
													Duration: game.DurationPermanent,
													Layer:    game.LayerPowerToughnessSet,
													SetPower: opt.Val(game.PT{
														Value: 4,
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
			{4}: This land becomes a 4/2 Golem artifact creature until end of turn. It's still a land.
		`,
		},
	}
}
