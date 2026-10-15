package p

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// PiracyCharm is the card definition for Piracy Charm.
//
// Type: Instant
// Cost: {U}
//
// Oracle text:
//
//	Choose one —
//	• Target creature gains islandwalk until end of turn. (It can't be blocked as long as defending player controls an Island.)
//	• Target creature gets +2/-1 until end of turn.
//	• Target player discards a card.
var PiracyCharm = newPiracyCharm

func newPiracyCharm() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue),
		CardFace: game.CardFace{
			Name: "Piracy Charm",
			ManaCost: opt.Val(cost.Mana{
				cost.U,
			}),
			Colors: []color.Color{color.Blue},
			Types:  []types.Card{types.Instant},
			SpellAbility: opt.Val(game.AbilityContent{
				Modes: []game.Mode{
					game.Mode{
						Text: "Target creature gains islandwalk until end of turn. (It can't be blocked as long as defending player controls an Island.)",
						Targets: []game.TargetSpec{
							game.TargetSpec{
								MinTargets: 1,
								MaxTargets: 1,
								Constraint: "target creature",
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
											Duration: game.DurationPermanent,
											Layer:    game.LayerAbility,
											AddAbilities: []game.Ability{
												&game.StaticAbility{
													Text: "Islandwalk",
													KeywordAbilities: []game.KeywordAbility{
														game.LandwalkKeyword{
															Subtype: types.Island,
														},
													},
												},
											},
										},
									},
									Duration: game.DurationUntilEndOfTurn,
								},
							},
						},
					},
					game.Mode{
						Text: "Target creature gets +2/-1 until end of turn.",
						Targets: []game.TargetSpec{
							game.TargetSpec{
								MinTargets: 1,
								MaxTargets: 1,
								Constraint: "target creature",
								Allow:      game.TargetAllowPermanent,
								Selection:  opt.Val(game.Selection{RequiredTypesAny: []types.Card{types.Creature}}),
							},
						},
						Sequence: []game.Instruction{
							{
								Primitive: game.ModifyPT{
									Object:         game.TargetPermanentReference(0),
									PowerDelta:     game.Fixed(2),
									ToughnessDelta: game.Fixed(-1),
									Duration:       game.DurationUntilEndOfTurn,
								},
							},
						},
					},
					game.Mode{
						Text: "Target player discards a card.",
						Targets: []game.TargetSpec{
							game.TargetSpec{
								MinTargets: 1,
								MaxTargets: 1,
								Constraint: "Target player",
								Allow:      game.TargetAllowPlayer,
							},
						},
						Sequence: []game.Instruction{
							{
								Primitive: game.Discard{
									Amount: game.Fixed(1),
									Player: game.TargetPlayerReference(0),
								},
							},
						},
					},
				},
				MinModes: 1,
				MaxModes: 1,
			}),
			OracleText: `
			Choose one —
			• Target creature gains islandwalk until end of turn. (It can't be blocked as long as defending player controls an Island.)
			• Target creature gets +2/-1 until end of turn.
			• Target player discards a card.
		`,
		},
	}
}
