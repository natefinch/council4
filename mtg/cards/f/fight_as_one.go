package f

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// FightAsOne is the card definition for Fight as One.
//
// Type: Instant
// Cost: {W}
//
// Oracle text:
//
//	Choose one or both —
//	• Target Human creature you control gets +1/+1 and gains indestructible until end of turn.
//	• Target non-Human creature you control gets +1/+1 and gains indestructible until end of turn.
var FightAsOne = newFightAsOne

func newFightAsOne() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White),
		CardFace: game.CardFace{
			Name: "Fight as One",
			ManaCost: opt.Val(cost.Mana{
				cost.W,
			}),
			Colors: []color.Color{color.White},
			Types:  []types.Card{types.Instant},
			SpellAbility: opt.Val(game.AbilityContent{
				Modes: []game.Mode{
					game.Mode{
						Text: "Target Human creature you control gets +1/+1 and gains indestructible until end of turn.",
						Targets: []game.TargetSpec{
							game.TargetSpec{
								MinTargets: 1,
								MaxTargets: 1,
								Constraint: "target Human creature you control",
								Allow:      game.TargetAllowPermanent,
								Selection:  opt.Val(game.Selection{RequiredTypesAny: []types.Card{types.Creature}, SubtypesAny: []types.Sub{types.Sub("Human")}, Controller: game.ControllerYou}),
							},
						},
						Sequence: []game.Instruction{
							{
								Primitive: game.ApplyContinuous{
									Object: opt.Val(game.TargetPermanentReference(0)),
									ContinuousEffects: []game.ContinuousEffect{
										game.ContinuousEffect{
											Layer:          game.LayerPowerToughnessModify,
											PowerDelta:     1,
											ToughnessDelta: 1,
										},
										game.ContinuousEffect{
											Layer: game.LayerAbility,
											AddKeywords: []game.Keyword{
												game.Indestructible,
											},
										},
									},
									Duration: game.DurationUntilEndOfTurn,
								},
							},
						},
					},
					game.Mode{
						Text: "Target non-Human creature you control gets +1/+1 and gains indestructible until end of turn.",
						Targets: []game.TargetSpec{
							game.TargetSpec{
								MinTargets: 1,
								MaxTargets: 1,
								Constraint: "target non-Human creature you control",
								Allow:      game.TargetAllowPermanent,
								Selection:  opt.Val(game.Selection{RequiredTypesAny: []types.Card{types.Creature}, ExcludedSubtype: types.Sub("Human"), Controller: game.ControllerYou}),
							},
						},
						Sequence: []game.Instruction{
							{
								Primitive: game.ApplyContinuous{
									Object: opt.Val(game.TargetPermanentReference(0)),
									ContinuousEffects: []game.ContinuousEffect{
										game.ContinuousEffect{
											Layer:          game.LayerPowerToughnessModify,
											PowerDelta:     1,
											ToughnessDelta: 1,
										},
										game.ContinuousEffect{
											Layer: game.LayerAbility,
											AddKeywords: []game.Keyword{
												game.Indestructible,
											},
										},
									},
									Duration: game.DurationUntilEndOfTurn,
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
			• Target Human creature you control gets +1/+1 and gains indestructible until end of turn.
			• Target non-Human creature you control gets +1/+1 and gains indestructible until end of turn.
		`,
		},
	}
}
