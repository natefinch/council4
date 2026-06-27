package h

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// Heal is the card definition for Heal.
//
// Type: Instant
// Cost: {W}
//
// Oracle text:
//
//	Prevent the next 1 damage that would be dealt to any target this turn.
//	Draw a card at the beginning of the next turn's upkeep.
var Heal = newHeal()

func newHeal() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White),
		CardFace: game.CardFace{
			Name: "Heal",
			ManaCost: opt.Val(cost.Mana{
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
							Amount:    game.Fixed(1),
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
			Prevent the next 1 damage that would be dealt to any target this turn.
			Draw a card at the beginning of the next turn's upkeep.
		`,
		},
	}
}
