package f

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// FacelessHaven is the card definition for Faceless Haven.
//
// Type: Snow Land
//
// Oracle text:
//
//	{T}: Add {C}.
//	{S}{S}{S}: This land becomes a 4/3 creature with vigilance and all creature types until end of turn. It's still a land. ({S} can be paid with one mana from a snow source.)
var FacelessHaven = newFacelessHaven

func newFacelessHaven() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name:       "Faceless Haven",
			Supertypes: []types.Super{types.Snow},
			Types:      []types.Card{types.Land},
			ActivatedAbilities: []game.ActivatedAbility{
				game.ActivatedAbility{
					Text: "{S}{S}{S}: This land becomes a 4/3 creature with vigilance and all creature types until end of turn. It's still a land. ({S} can be paid with one mana from a snow source.)",
					ManaCost: opt.Val(cost.Mana{
						cost.S,
						cost.S,
						cost.S,
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
													Duration:    game.DurationPermanent,
													Layer:       game.LayerAbility,
													AddKeywords: []game.Keyword{game.Vigilance},
												},
												game.ContinuousEffect{
													Duration: game.DurationPermanent,
													Layer:    game.LayerPowerToughnessSet,
													SetPower: opt.Val(game.PT{
														Value: 4,
													}),
													SetToughness: opt.Val(game.PT{
														Value: 3,
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
			{S}{S}{S}: This land becomes a 4/3 creature with vigilance and all creature types until end of turn. It's still a land. ({S} can be paid with one mana from a snow source.)
		`,
		},
	}
}
