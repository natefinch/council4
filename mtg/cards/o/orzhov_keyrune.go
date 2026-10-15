package o

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// OrzhovKeyrune is the card definition for Orzhov Keyrune.
//
// Type: Artifact
// Cost: {3}
//
// Oracle text:
//
//	{T}: Add {W} or {B}.
//	{W}{B}: This artifact becomes a 1/4 white and black Thrull artifact creature with lifelink until end of turn.
var OrzhovKeyrune = newOrzhovKeyrune

func newOrzhovKeyrune() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White, color.Black),
		CardFace: game.CardFace{
			Name: "Orzhov Keyrune",
			ManaCost: opt.Val(cost.Mana{
				cost.O(3),
			}),
			Types: []types.Card{types.Artifact},
			ActivatedAbilities: []game.ActivatedAbility{
				game.ActivatedAbility{
					Text: "{W}{B}: This artifact becomes a 1/4 white and black Thrull artifact creature with lifelink until end of turn.",
					ManaCost: opt.Val(cost.Mana{
						cost.W,
						cost.B,
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
													Duration:  game.DurationPermanent,
													Layer:     game.LayerColor,
													SetColors: []color.Color{color.White, color.Black},
												},
												game.ContinuousEffect{
													Duration:    game.DurationPermanent,
													Layer:       game.LayerType,
													AddTypes:    []types.Card{types.Creature, types.Artifact},
													AddSubtypes: []types.Sub{types.Thrull},
												},
												game.ContinuousEffect{
													Duration:    game.DurationPermanent,
													Layer:       game.LayerAbility,
													AddKeywords: []game.Keyword{game.Lifelink},
												},
												game.ContinuousEffect{
													Duration: game.DurationPermanent,
													Layer:    game.LayerPowerToughnessSet,
													SetPower: opt.Val(game.PT{
														Value: 1,
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
				game.TapManaChoiceAbility(mana.W, mana.B),
			},
			OracleText: `
			{T}: Add {W} or {B}.
			{W}{B}: This artifact becomes a 1/4 white and black Thrull artifact creature with lifelink until end of turn.
		`,
		},
	}
}
