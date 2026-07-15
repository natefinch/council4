package n

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// Necropotence is the card definition for Necropotence.
//
// Type: Enchantment
// Cost: {B}{B}{B}
//
// Oracle text:
//
//	Skip your draw step.
//	Whenever you discard a card, exile that card from your graveyard.
//	Pay 1 life: Exile the top card of your library face down. Put that card into your hand at the beginning of your next end step.
var Necropotence = newNecropotence

func newNecropotence() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Black),
		CardFace: game.CardFace{
			Name: "Necropotence",
			ManaCost: opt.Val(cost.Mana{
				cost.B,
				cost.B,
				cost.B,
			}),
			Colors: []color.Color{color.Black},
			Types:  []types.Card{types.Enchantment},
			StaticAbilities: []game.StaticAbility{
				game.SkipDrawStepStaticBody,
			},
			ActivatedAbilities: []game.ActivatedAbility{
				game.ActivatedAbility{
					Text: "Pay 1 life: Exile the top card of your library face down. Put that card into your hand at the beginning of your next end step.",
					AdditionalCosts: []cost.Additional{
						{
							Kind:   cost.AdditionalPayLife,
							Text:   "Pay 1 life",
							Amount: 1,
						},
					},
					ZoneOfFunction: zone.Battlefield,
					Content: game.Mode{
						Sequence: []game.Instruction{
							{
								Primitive: game.ExileTopOfLibrary{
									Amount:        game.Fixed(1),
									Player:        game.ControllerReference(),
									PublishLinked: game.LinkedKey("delayed-top-card-1"),
									FaceDown:      true,
								},
							},
							{
								Primitive: game.CreateDelayedTrigger{
									Trigger: game.DelayedTriggerDef{
										Timing:       game.DelayedAtBeginningOfYourNextEndStep,
										CapturedCard: opt.Val(game.LinkedObjectReference("delayed-top-card-1")),
										Content: game.Mode{
											Sequence: []game.Instruction{
												{
													Primitive: game.MoveCard{
														Card:        game.CardReference{Kind: game.CardReferenceCaptured},
														FromZone:    zone.Exile,
														Destination: zone.Hand,
													},
												},
											},
										}.Ability(),
									},
								},
							},
						},
					}.Ability(),
				},
			},
			TriggeredAbilities: []game.TriggeredAbility{
				game.TriggeredAbility{
					Trigger: game.TriggerCondition{
						Type: game.TriggerWhenever,
						Pattern: game.TriggerPattern{
							Event:  game.EventCardDiscarded,
							Player: game.TriggerPlayerYou,
						},
					},
					Content: game.Mode{
						Sequence: []game.Instruction{
							{
								Primitive: game.MoveCard{
									Card:        game.CardReference{Kind: game.CardReferenceEvent},
									FromZone:    zone.Graveyard,
									Destination: zone.Exile,
								},
							},
						},
					}.Ability(),
				},
			},
			OracleText: `
			Skip your draw step.
			Whenever you discard a card, exile that card from your graveyard.
			Pay 1 life: Exile the top card of your library face down. Put that card into your hand at the beginning of your next end step.
		`,
		},
	}
}
