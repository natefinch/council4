package m

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// Memnarch is the card definition for Memnarch.
//
// Type: Legendary Artifact Creature — Wizard
// Cost: {7}
//
// Oracle text:
//
//	{1}{U}{U}: Target permanent becomes an artifact in addition to its other types. (This effect lasts indefinitely.)
//	{3}{U}: Gain control of target artifact. (This effect lasts indefinitely.)
var Memnarch = newMemnarch

func newMemnarch() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue),
		CardFace: game.CardFace{
			Name: "Memnarch",
			ManaCost: opt.Val(cost.Mana{
				cost.O(7),
			}),
			Supertypes: []types.Super{types.Legendary},
			Types:      []types.Card{types.Artifact, types.Creature},
			Subtypes:   []types.Sub{types.Wizard},
			Power:      opt.Val(game.PT{Value: 4}),
			Toughness:  opt.Val(game.PT{Value: 5}),
			ActivatedAbilities: []game.ActivatedAbility{
				game.ActivatedAbility{
					Text: "{1}{U}{U}: Target permanent becomes an artifact in addition to its other types. (This effect lasts indefinitely.)",
					ManaCost: opt.Val(cost.Mana{
						cost.O(1),
						cost.U,
						cost.U,
					}),
					ZoneOfFunction: zone.Battlefield,
					Content: game.AbilityContent{
						Modes: []game.Mode{
							game.Mode{
								Targets: []game.TargetSpec{
									game.TargetSpec{
										MinTargets: 1,
										MaxTargets: 1,
										Constraint: "target permanent",
										Allow:      game.TargetAllowPermanent,
									},
								},
								Sequence: []game.Instruction{
									game.Instruction{
										Primitive: game.ApplyContinuous{
											Object: opt.Val(game.TargetPermanentReference(0)),
											ContinuousEffects: []game.ContinuousEffect{
												game.ContinuousEffect{
													Duration: game.DurationPermanent,
													Layer:    game.LayerType,
													AddTypes: []types.Card{types.Artifact},
												},
											},
											Duration: game.DurationPermanent,
										},
									},
								},
							},
						},
						MinModes: 1,
						MaxModes: 1,
					},
				},
				game.ActivatedAbility{
					Text: "{3}{U}: Gain control of target artifact. (This effect lasts indefinitely.)",
					ManaCost: opt.Val(cost.Mana{
						cost.O(3),
						cost.U,
					}),
					ZoneOfFunction: zone.Battlefield,
					Content: game.AbilityContent{
						Modes: []game.Mode{
							game.Mode{
								Targets: []game.TargetSpec{
									game.TargetSpec{
										MinTargets: 1,
										MaxTargets: 1,
										Constraint: "target artifact",
										Allow:      game.TargetAllowPermanent,
										Selection:  opt.Val(game.Selection{RequiredTypesAny: []types.Card{types.Artifact}}),
									},
								},
								Sequence: []game.Instruction{
									game.Instruction{
										Primitive: game.ApplyContinuous{
											Object: opt.Val(game.TargetPermanentReference(0)),
											ContinuousEffects: []game.ContinuousEffect{
												game.ContinuousEffect{
													Duration:      game.DurationPermanent,
													Layer:         game.LayerControl,
													NewController: opt.Val(game.Player1),
												},
											},
											Duration: game.DurationPermanent,
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
			{1}{U}{U}: Target permanent becomes an artifact in addition to its other types. (This effect lasts indefinitely.)
			{3}{U}: Gain control of target artifact. (This effect lasts indefinitely.)
		`,
		},
	}
}
