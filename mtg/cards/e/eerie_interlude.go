package e

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// EerieInterlude is the card definition for Eerie Interlude.
//
// Type: Instant
// Cost: {2}{W}
//
// Oracle text:
//
//	Exile any number of target creatures you control. Return those cards to the battlefield under their owner's control at the beginning of the next end step.
var EerieInterlude = newEerieInterlude

func newEerieInterlude() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White),
		CardFace: game.CardFace{
			Name: "Eerie Interlude",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
				cost.W,
			}),
			Colors: []color.Color{color.White},
			Types:  []types.Card{types.Instant},
			SpellAbility: opt.Val(game.Mode{
				Targets: []game.TargetSpec{
					game.TargetSpec{
						MinTargets: 0,
						MaxTargets: 99,
						Constraint: "any number of target creatures you control",
						Allow:      game.TargetAllowPermanent,
						Selection:  opt.Val(game.Selection{RequiredTypesAny: []types.Card{types.Creature}, Controller: game.ControllerYou}),
					},
				},
				Sequence: []game.Instruction{
					{
						Primitive: game.Exile{
							Object:         game.AllTargetPermanentsReference(0),
							ExileLinkedKey: game.LinkedKey("group-blink"),
						},
					},
					{
						Primitive: game.CreateDelayedTrigger{
							Trigger: game.DelayedTriggerDef{
								Timing: game.DelayedAtBeginningOfNextEndStep,
								Content: game.Mode{
									Sequence: []game.Instruction{
										{
											Primitive: game.PutOnBattlefield{
												Source: game.LinkedBattlefieldSource(game.LinkedKey("group-blink")),
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
			Exile any number of target creatures you control. Return those cards to the battlefield under their owner's control at the beginning of the next end step.
		`,
		},
	}
}
