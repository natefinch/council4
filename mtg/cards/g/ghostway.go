package g

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// Ghostway is the card definition for Ghostway.
//
// Type: Instant
// Cost: {2}{W}
//
// Oracle text:
//
//	Exile each creature you control. Return those cards to the battlefield under their owner's control at the beginning of the next end step.
var Ghostway = newGhostway

func newGhostway() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White),
		CardFace: game.CardFace{
			Name: "Ghostway",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
				cost.W,
			}),
			Colors: []color.Color{color.White},
			Types:  []types.Card{types.Instant},
			SpellAbility: opt.Val(game.Mode{
				Sequence: []game.Instruction{
					{
						Primitive: game.Exile{
							Group:          game.BattlefieldGroup(game.Selection{RequiredTypes: []types.Card{types.Creature}, Controller: game.ControllerYou}),
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
			Exile each creature you control. Return those cards to the battlefield under their owner's control at the beginning of the next end step.
		`,
		},
	}
}
