package b

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// BorrowedHostility is the card definition for Borrowed Hostility.
//
// Type: Instant
// Cost: {R}
//
// Oracle text:
//
//	Escalate {3} (Pay this cost for each mode chosen beyond the first.)
//	Choose one or both —
//	• Target creature gets +3/+0 until end of turn.
//	• Target creature gains first strike until end of turn.
var BorrowedHostility = newBorrowedHostility

func newBorrowedHostility() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Red),
		CardFace: game.CardFace{
			Name: "Borrowed Hostility",
			ManaCost: opt.Val(cost.Mana{
				cost.R,
			}),
			Colors: []color.Color{color.Red},
			Types:  []types.Card{types.Instant},
			SpellAbility: opt.Val(game.AbilityContent{
				Modes: []game.Mode{
					game.Mode{
						Text: "Target creature gets +3/+0 until end of turn.",
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
									Object:     game.TargetPermanentReference(0),
									PowerDelta: game.Fixed(3),
									Duration:   game.DurationUntilEndOfTurn,
								},
							},
						},
					},
					game.Mode{
						Text: "Target creature gains first strike until end of turn.",
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
											Duration:    game.DurationPermanent,
											Layer:       game.LayerAbility,
											AddKeywords: []game.Keyword{game.FirstStrike},
										},
									},
									Duration: game.DurationUntilEndOfTurn,
								},
							},
						},
					},
				},
				MinModes:     1,
				MaxModes:     2,
				EscalateCost: opt.Val(cost.Mana{cost.O(3)}),
			}),
			OracleText: `
			Escalate {3} (Pay this cost for each mode chosen beyond the first.)
			Choose one or both —
			• Target creature gets +3/+0 until end of turn.
			• Target creature gains first strike until end of turn.
		`,
		},
	}
}
