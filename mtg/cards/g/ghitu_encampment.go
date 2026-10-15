package g

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// GhituEncampment is the card definition for Ghitu Encampment.
//
// Type: Land
//
// Oracle text:
//
//	This land enters tapped.
//	{T}: Add {R}.
//	{1}{R}: This land becomes a 2/1 red Warrior creature with first strike until end of turn. It's still a land. (It deals combat damage before creatures without first strike.)
var GhituEncampment = newGhituEncampment

func newGhituEncampment() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Red),
		CardFace: game.CardFace{
			Name:  "Ghitu Encampment",
			Types: []types.Card{types.Land},
			ActivatedAbilities: []game.ActivatedAbility{
				game.ActivatedAbility{
					Text: "{1}{R}: This land becomes a 2/1 red Warrior creature with first strike until end of turn. It's still a land. (It deals combat damage before creatures without first strike.)",
					ManaCost: opt.Val(cost.Mana{
						cost.O(1),
						cost.R,
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
													SetColors: []color.Color{color.Red},
												},
												game.ContinuousEffect{
													Duration:    game.DurationPermanent,
													Layer:       game.LayerType,
													AddTypes:    []types.Card{types.Creature},
													AddSubtypes: []types.Sub{types.Warrior},
												},
												game.ContinuousEffect{
													Duration:    game.DurationPermanent,
													Layer:       game.LayerAbility,
													AddKeywords: []game.Keyword{game.FirstStrike},
												},
												game.ContinuousEffect{
													Duration: game.DurationPermanent,
													Layer:    game.LayerPowerToughnessSet,
													SetPower: opt.Val(game.PT{
														Value: 2,
													}),
													SetToughness: opt.Val(game.PT{
														Value: 1,
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
				game.TapManaAbility(mana.R),
			},
			ReplacementAbilities: []game.ReplacementAbility{
				game.EntersTappedReplacement("This land enters tapped."),
			},
			OracleText: `
			This land enters tapped.
			{T}: Add {R}.
			{1}{R}: This land becomes a 2/1 red Warrior creature with first strike until end of turn. It's still a land. (It deals combat damage before creatures without first strike.)
		`,
		},
	}
}
