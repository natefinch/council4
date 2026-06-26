package t

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// TabletOfEpityr is the card definition for Tablet of Epityr.
//
// Type: Artifact
// Cost: {1}
//
// Oracle text:
//
//	Whenever an artifact you control is put into a graveyard from the battlefield, you may pay {1}. If you do, you gain 1 life.
var TabletOfEpityr = newTabletOfEpityr()

func newTabletOfEpityr() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name: "Tablet of Epityr",
			ManaCost: opt.Val(cost.Mana{
				cost.O(1),
			}),
			Types: []types.Card{types.Artifact},
			TriggeredAbilities: []game.TriggeredAbility{
				game.TriggeredAbility{
					Trigger: game.TriggerCondition{
						Type: game.TriggerWhenever,
						Pattern: game.TriggerPattern{
							Event:            game.EventZoneChanged,
							Controller:       game.TriggerControllerYou,
							MatchFromZone:    true,
							FromZone:         zone.Battlefield,
							MatchToZone:      true,
							ToZone:           zone.Graveyard,
							SubjectSelection: game.Selection{RequiredTypes: []types.Card{types.Artifact}},
						},
					},
					Content: game.Mode{
						Sequence: []game.Instruction{
							{
								Primitive: game.Pay{
									Payment: game.ResolutionPayment{
										Prompt: "Pay {1}?",
										ManaCost: opt.Val(cost.Mana{
											cost.O(1),
										}),
									},
								},
								PublishResult: game.ResultKey("controller-paid"),
							},
							{
								Primitive: game.GainLife{
									Amount: game.Fixed(1),
									Player: game.ControllerReference(),
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
			Whenever an artifact you control is put into a graveyard from the battlefield, you may pay {1}. If you do, you gain 1 life.
		`,
		},
	}
}
