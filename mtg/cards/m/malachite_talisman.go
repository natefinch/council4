package m

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// MalachiteTalisman is the card definition for Malachite Talisman.
//
// Type: Artifact
// Cost: {2}
//
// Oracle text:
//
//	Whenever a player casts a green spell, you may pay {3}. If you do, untap target permanent.
var MalachiteTalisman = newMalachiteTalisman()

func newMalachiteTalisman() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name: "Malachite Talisman",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
			}),
			Types: []types.Card{types.Artifact},
			TriggeredAbilities: []game.TriggeredAbility{
				game.TriggeredAbility{
					Trigger: game.TriggerCondition{
						Type: game.TriggerWhenever,
						Pattern: game.TriggerPattern{
							Event:         game.EventSpellCast,
							CardSelection: game.Selection{ColorsAny: []color.Color{color.Green}},
						},
					},
					Content: game.Mode{
						Targets: []game.TargetSpec{
							game.TargetSpec{
								MinTargets: 1,
								MaxTargets: 1,
								Constraint: "target permanent",
								Allow:      game.TargetAllowPermanent,
							},
						},
						Sequence: []game.Instruction{
							{
								Primitive: game.Pay{
									Payment: game.ResolutionPayment{
										Prompt: "Pay {3}?",
										ManaCost: opt.Val(cost.Mana{
											cost.O(3),
										}),
									},
								},
								PublishResult: game.ResultKey("controller-paid"),
							},
							{
								Primitive: game.Untap{
									Object: game.TargetPermanentReference(0),
								},
								ResultGate: opt.Val(game.InstructionResultGate{
									Key:       "controller-paid",
									Succeeded: game.TriTrue,
								}),
							},
						},
					}.Ability(),
				},
			},
			OracleText: `
			Whenever a player casts a green spell, you may pay {3}. If you do, untap target permanent.
		`,
		},
	}
}
