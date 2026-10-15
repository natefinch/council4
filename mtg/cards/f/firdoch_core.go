package f

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// FirdochCore is the card definition for Firdoch Core.
//
// Type: Kindred Artifact — Shapeshifter
// Cost: {3}
//
// Oracle text:
//
//	Changeling (This card is every creature type.)
//	{T}: Add one mana of any color.
//	{4}: This artifact becomes a 4/4 artifact creature until end of turn.
var FirdochCore = newFirdochCore

func newFirdochCore() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name: "Firdoch Core",
			ManaCost: opt.Val(cost.Mana{
				cost.O(3),
			}),
			Types:    []types.Card{types.Kindred, types.Artifact},
			Subtypes: []types.Sub{types.Shapeshifter},
			StaticAbilities: []game.StaticAbility{
				game.ChangelingStaticBody,
			},
			ActivatedAbilities: []game.ActivatedAbility{
				game.ActivatedAbility{
					Text: "{4}: This artifact becomes a 4/4 artifact creature until end of turn.",
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
													Duration: game.DurationPermanent,
													Layer:    game.LayerType,
													AddTypes: []types.Card{types.Creature, types.Artifact},
												},
												game.ContinuousEffect{
													Duration: game.DurationPermanent,
													Layer:    game.LayerPowerToughnessSet,
													SetPower: opt.Val(game.PT{
														Value: 4,
													}),
													SetToughness: opt.Val(game.PT{
														Value: 4,
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
				game.TapManaChoiceAbility(mana.W, mana.U, mana.B, mana.R, mana.G),
			},
			OracleText: `
			Changeling (This card is every creature type.)
			{T}: Add one mana of any color.
			{4}: This artifact becomes a 4/4 artifact creature until end of turn.
		`,
		},
	}
}
