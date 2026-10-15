package b

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// BorrowedGrace is the card definition for Borrowed Grace.
//
// Type: Instant
// Cost: {2}{W}
//
// Oracle text:
//
//	Escalate {1}{W} (Pay this cost for each mode chosen beyond the first.)
//	Choose one or both —
//	• Creatures you control get +2/+0 until end of turn.
//	• Creatures you control get +0/+2 until end of turn.
var BorrowedGrace = newBorrowedGrace

func newBorrowedGrace() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White),
		CardFace: game.CardFace{
			Name: "Borrowed Grace",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
				cost.W,
			}),
			Colors: []color.Color{color.White},
			Types:  []types.Card{types.Instant},
			SpellAbility: opt.Val(game.AbilityContent{
				Modes: []game.Mode{
					game.Mode{
						Text: "Creatures you control get +2/+0 until end of turn.",
						Sequence: []game.Instruction{
							{
								Primitive: game.ApplyContinuous{
									ContinuousEffects: []game.ContinuousEffect{
										game.ContinuousEffect{
											Duration:   game.DurationPermanent,
											Group:      game.BattlefieldGroup(game.Selection{RequiredTypes: []types.Card{types.Creature}, Controller: game.ControllerYou}),
											Layer:      game.LayerPowerToughnessModify,
											PowerDelta: 2,
										},
									},
									Duration: game.DurationUntilEndOfTurn,
								},
							},
						},
					},
					game.Mode{
						Text: "Creatures you control get +0/+2 until end of turn.",
						Sequence: []game.Instruction{
							{
								Primitive: game.ApplyContinuous{
									ContinuousEffects: []game.ContinuousEffect{
										game.ContinuousEffect{
											Duration:       game.DurationPermanent,
											Group:          game.BattlefieldGroup(game.Selection{RequiredTypes: []types.Card{types.Creature}, Controller: game.ControllerYou}),
											Layer:          game.LayerPowerToughnessModify,
											ToughnessDelta: 2,
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
				EscalateCost: opt.Val(cost.Mana{cost.O(1), cost.W}),
			}),
			OracleText: `
			Escalate {1}{W} (Pay this cost for each mode chosen beyond the first.)
			Choose one or both —
			• Creatures you control get +2/+0 until end of turn.
			• Creatures you control get +0/+2 until end of turn.
		`,
		},
	}
}
