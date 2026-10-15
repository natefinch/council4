package c

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// ChimericIdol is the card definition for Chimeric Idol.
//
// Type: Artifact
// Cost: {3}
//
// Oracle text:
//
//	{0}: Tap all lands you control. This artifact becomes a 3/3 Turtle artifact creature until end of turn.
var ChimericIdol = newChimericIdol

func newChimericIdol() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name: "Chimeric Idol",
			ManaCost: opt.Val(cost.Mana{
				cost.O(3),
			}),
			Types: []types.Card{types.Artifact},
			ActivatedAbilities: []game.ActivatedAbility{
				game.ActivatedAbility{
					Text: "{0}: Tap all lands you control. This artifact becomes a 3/3 Turtle artifact creature until end of turn.",
					ManaCost: opt.Val(cost.Mana{
						cost.O(0),
					}),
					ZoneOfFunction: zone.Battlefield,
					Content: game.AbilityContent{
						Modes: []game.Mode{
							game.Mode{
								Sequence: []game.Instruction{
									game.Instruction{
										Primitive: game.Tap{
											Group: game.BattlefieldGroup(game.Selection{RequiredTypes: []types.Card{types.Land}, Controller: game.ControllerYou}),
										},
									},
									game.Instruction{
										Primitive: game.ApplyContinuous{
											Object: opt.Val(game.SourcePermanentReference()),
											ContinuousEffects: []game.ContinuousEffect{
												game.ContinuousEffect{
													Duration:    game.DurationPermanent,
													Layer:       game.LayerType,
													AddTypes:    []types.Card{types.Creature, types.Artifact},
													AddSubtypes: []types.Sub{types.Turtle},
												},
												game.ContinuousEffect{
													Duration: game.DurationPermanent,
													Layer:    game.LayerPowerToughnessSet,
													SetPower: opt.Val(game.PT{
														Value: 3,
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
			OracleText: `
			{0}: Tap all lands you control. This artifact becomes a 3/3 Turtle artifact creature until end of turn.
		`,
		},
	}
}
