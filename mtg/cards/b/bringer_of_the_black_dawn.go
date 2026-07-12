package b

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// BringerOfTheBlackDawn is the card definition for Bringer of the Black Dawn.
//
// Type: Creature — Bringer
// Cost: {7}{B}{B}
//
// Oracle text:
//
//	You may pay {W}{U}{B}{R}{G} rather than pay this spell's mana cost.
//	Trample
//	At the beginning of your upkeep, you may pay 2 life. If you do, search your library for a card, then shuffle and put that card on top.
var BringerOfTheBlackDawn = newBringerOfTheBlackDawn

func newBringerOfTheBlackDawn() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White, color.Blue, color.Black, color.Red, color.Green),
		CardFace: game.CardFace{
			Name: "Bringer of the Black Dawn",
			ManaCost: opt.Val(cost.Mana{
				cost.O(7),
				cost.B,
				cost.B,
			}),
			Colors:    []color.Color{color.Black},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Bringer},
			Power:     opt.Val(game.PT{Value: 5}),
			Toughness: opt.Val(game.PT{Value: 5}),
			StaticAbilities: []game.StaticAbility{
				game.TrampleStaticBody,
			},
			TriggeredAbilities: []game.TriggeredAbility{
				game.TriggeredAbility{
					Trigger: game.TriggerCondition{
						Type: game.TriggerAt,
						Pattern: game.TriggerPattern{
							Event:      game.EventBeginningOfStep,
							Controller: game.TriggerControllerYou,
							Step:       game.StepUpkeep,
						},
					},
					Content: game.Mode{
						Sequence: []game.Instruction{
							{
								Primitive: game.Pay{
									Payment: game.ResolutionPayment{
										Prompt: "Pay 2 life?",
										AdditionalCosts: []cost.Additional{
											{
												Kind:   cost.AdditionalPayLife,
												Text:   "pay 2 life",
												Amount: 2,
											},
										},
									},
								},
								PublishResult: game.ResultKey("controller-paid"),
							},
							{
								Primitive: game.Search{
									Player: game.ControllerReference(),
									Spec: game.SearchSpec{
										SourceZone:          zone.Library,
										Destination:         zone.Library,
										DestinationPosition: game.SearchPositionTop,
										FailToFindPolicy:    game.SearchMustFindIfAvailable,
									},
									Amount: game.Fixed(1),
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
			AlternativeCosts: []cost.Alternative{
				cost.Alternative{
					Label:    "Pay {W}{U}{B}{R}{G}",
					ManaCost: opt.Val(cost.Mana{cost.W, cost.U, cost.B, cost.R, cost.G}),
				},
			},
			OracleText: `
			You may pay {W}{U}{B}{R}{G} rather than pay this spell's mana cost.
			Trample
			At the beginning of your upkeep, you may pay 2 life. If you do, search your library for a card, then shuffle and put that card on top.
		`,
		},
	}
}
