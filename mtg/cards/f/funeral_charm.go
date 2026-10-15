package f

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// FuneralCharm is the card definition for Funeral Charm.
//
// Type: Instant
// Cost: {B}
//
// Oracle text:
//
//	Choose one —
//	• Target player discards a card.
//	• Target creature gets +2/-1 until end of turn.
//	• Target creature gains swampwalk until end of turn. (It can't be blocked as long as defending player controls a Swamp.)
var FuneralCharm = newFuneralCharm

func newFuneralCharm() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Black),
		CardFace: game.CardFace{
			Name: "Funeral Charm",
			ManaCost: opt.Val(cost.Mana{
				cost.B,
			}),
			Colors: []color.Color{color.Black},
			Types:  []types.Card{types.Instant},
			SpellAbility: opt.Val(game.AbilityContent{
				Modes: []game.Mode{
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
						Text: "Target creature gains swampwalk until end of turn. (It can't be blocked as long as defending player controls a Swamp.)",
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
													Text: "Swampwalk",
													KeywordAbilities: []game.KeywordAbility{
														game.LandwalkKeyword{
															Subtype: types.Swamp,
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
				},
				MinModes: 1,
				MaxModes: 1,
			}),
			OracleText: `
			Choose one —
			• Target player discards a card.
			• Target creature gets +2/-1 until end of turn.
			• Target creature gains swampwalk until end of turn. (It can't be blocked as long as defending player controls a Swamp.)
		`,
		},
	}
}
