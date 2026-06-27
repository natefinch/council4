package s

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// SwiftManeuver is the card definition for Swift Maneuver.
//
// Type: Instant
// Cost: {1}{W}
//
// Oracle text:
//
//	Prevent the next 2 damage that would be dealt to any target this turn.
//	Draw a card at the beginning of the next turn's upkeep.
var SwiftManeuver = newSwiftManeuver()

func newSwiftManeuver() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White),
		CardFace: game.CardFace{
			Name: "Swift Maneuver",
			ManaCost: opt.Val(cost.Mana{
				cost.O(1),
				cost.W,
			}),
			Colors: []color.Color{color.White},
			Types:  []types.Card{types.Instant},
			SpellAbility: opt.Val(game.Mode{
				Targets: []game.TargetSpec{
					game.TargetSpec{
						MinTargets: 1,
						MaxTargets: 1,
						Constraint: "any target",
						Allow:      game.TargetAllowPermanent | game.TargetAllowPlayer,
					},
				},
				Sequence: []game.Instruction{
					{
						Primitive: game.PreventDamage{
							AnyTarget: game.AnyTargetDamageRecipient(0),
							Amount:    game.Fixed(2),
						},
					},
					{
						Primitive: game.CreateDelayedTrigger{
							Trigger: game.DelayedTriggerDef{
								Timing: game.DelayedAtBeginningOfNextUpkeep,
								Content: game.Mode{
									Sequence: []game.Instruction{
										{
											Primitive: game.Draw{
												Amount: game.Fixed(1),
												Player: game.ControllerReference(),
											},
										},
									},
								}.Ability(),
							},
						},
					},
				},
			}.Ability()),
			OracleText: `
			Prevent the next 2 damage that would be dealt to any target this turn.
			Draw a card at the beginning of the next turn's upkeep.
		`,
		},
	}
}
