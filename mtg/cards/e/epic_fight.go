package e

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// EpicFight is the card definition for Epic Fight.
//
// Type: Sorcery
// Cost: {2}{G}
//
// Oracle text:
//
//	Choose one or both —
//	• Double target creature's power and toughness until end of turn.
//	• Target creature you control fights target creature an opponent controls.
var EpicFight = newEpicFight()

func newEpicFight() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Green),
		CardFace: game.CardFace{
			Name: "Epic Fight",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
				cost.G,
			}),
			Colors: []color.Color{color.Green},
			Types:  []types.Card{types.Sorcery},
			SpellAbility: opt.Val(game.AbilityContent{
				Modes: []game.Mode{
					game.Mode{
						Text: "Double target creature's power and toughness until end of turn.",
						Targets: []game.TargetSpec{
							game.TargetSpec{
								MinTargets: 1,
								MaxTargets: 1,
								Constraint: "target creature's power and toughness",
								Allow:      game.TargetAllowPermanent,
								Selection:  opt.Val(game.Selection{RequiredTypesAny: []types.Card{types.Creature}}),
							},
						},
						Sequence: []game.Instruction{
							{
								Primitive: game.ApplyContinuous{
									Object: opt.Val(game.TargetPermanentReference(0)),
									ContinuousEffects: []game.ContinuousEffect{
										game.ContinuousEffect{
											Layer:           game.LayerPowerToughnessModify,
											DoublePower:     true,
											DoubleToughness: true,
										},
									},
									Duration: game.DurationUntilEndOfTurn,
								},
							},
						},
					},
					game.Mode{
						Text: "Target creature you control fights target creature an opponent controls.",
						Targets: []game.TargetSpec{
							game.TargetSpec{
								MinTargets: 1,
								MaxTargets: 1,
								Constraint: "Target creature you control",
								Allow:      game.TargetAllowPermanent,
								Selection:  opt.Val(game.Selection{RequiredTypesAny: []types.Card{types.Creature}, Controller: game.ControllerYou}),
							},
							game.TargetSpec{
								MinTargets: 1,
								MaxTargets: 1,
								Constraint: "target creature an opponent controls",
								Allow:      game.TargetAllowPermanent,
								Selection:  opt.Val(game.Selection{RequiredTypesAny: []types.Card{types.Creature}, Controller: game.ControllerOpponent}),
							},
						},
						Sequence: []game.Instruction{
							{
								Primitive: game.Fight{
									Object:        game.TargetPermanentReference(0),
									RelatedObject: game.TargetPermanentReference(1),
								},
							},
						},
					},
				},
				MinModes: 1,
				MaxModes: 2,
			}),
			OracleText: `
			Choose one or both —
			• Double target creature's power and toughness until end of turn.
			• Target creature you control fights target creature an opponent controls.
		`,
		},
	}
}
