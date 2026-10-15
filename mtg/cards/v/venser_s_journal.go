package v

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// VenserSJournal is the card definition for Venser's Journal.
//
// Type: Artifact — Book
// Cost: {5}
//
// Oracle text:
//
//	You have no maximum hand size.
//	At the beginning of your upkeep, you gain 1 life for each card in your hand.
var VenserSJournal = newVenserSJournal

func newVenserSJournal() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name: "Venser's Journal",
			ManaCost: opt.Val(cost.Mana{
				cost.O(5),
			}),
			Types:    []types.Card{types.Artifact},
			Subtypes: []types.Sub{types.Book},
			StaticAbilities: []game.StaticAbility{
				game.NoMaximumHandSizeStaticBody,
			},
			TriggeredAbilities: []game.TriggeredAbility{
				game.TriggeredAbility{
					Text: "At the beginning of your upkeep, you gain 1 life for each card in your hand.",
					Trigger: game.TriggerCondition{
						Type: game.TriggerAt,
						Pattern: game.TriggerPattern{
							Event:           game.EventBeginningOfStep,
							Controller:      game.TriggerControllerYou,
							StackObjectKind: game.StackSpell,
							CounterKind:     counter.PlusOnePlusOne,
							Step:            game.StepUpkeep,
						},
					},
					Content: game.AbilityContent{
						Modes: []game.Mode{
							game.Mode{
								Sequence: []game.Instruction{
									game.Instruction{
										Primitive: game.GainLife{
											Amount: game.Dynamic(game.DynamicAmount{
												Kind:       game.DynamicAmountCountCardsInZone,
												Multiplier: 1,
												Player:     func() *game.PlayerReference { ref := game.ControllerReference(); return &ref }(),
												CardZone:   zone.Hand,
												Selection:  &game.Selection{},
											}),
											Player: game.ControllerReference(),
										},
									},
								},
							},
						},
						MinModes: 1,
						MaxModes: 1,
					},
				},
			},
			OracleText: `
			You have no maximum hand size.
			At the beginning of your upkeep, you gain 1 life for each card in your hand.
		`,
		},
	}
}
