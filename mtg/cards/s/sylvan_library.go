package s

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// SylvanLibrary is the card definition for Sylvan Library.
//
// Type: Enchantment
// Cost: {1}{G}
//
// Oracle text:
//
//	At the beginning of your draw step, you may draw two additional cards. If you do, choose two cards in your hand drawn this turn. For each of those cards, pay 4 life or put the card on top of your library.
var SylvanLibrary = newSylvanLibrary

func newSylvanLibrary() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Green),
		CardFace: game.CardFace{
			Name: "Sylvan Library",
			ManaCost: opt.Val(cost.Mana{
				cost.O(1),
				cost.G,
			}),
			Colors: []color.Color{color.Green},
			Types:  []types.Card{types.Enchantment},
			TriggeredAbilities: []game.TriggeredAbility{
				game.TriggeredAbility{
					Trigger: game.TriggerCondition{
						Type: game.TriggerAt,
						Pattern: game.TriggerPattern{
							Event:      game.EventBeginningOfStep,
							Controller: game.TriggerControllerYou,
							Step:       game.StepDraw,
						},
					},
					Content: game.Mode{
						Text: "At the beginning of your draw step, you may draw two additional cards. If you do, choose two cards in your hand drawn this turn. For each of those cards, pay 4 life or put the card on top of your library.",
						Sequence: []game.Instruction{
							{
								Primitive: game.Draw{
									Amount: game.Fixed(2),
									Player: game.ControllerReference(),
								},
								Optional:      true,
								PublishResult: game.ResultKey("sylvan-extra-draw-drew"),
							},
							{
								Primitive: game.ChooseDrawnPayLifeOrTop{
									Player:      game.ControllerReference(),
									ChooseCount: 2,
									LifeCost:    4,
								},
								ResultGate: opt.Val(game.InstructionResultGate{
									Key:      "sylvan-extra-draw-drew",
									Accepted: game.TriTrue,
								}),
							},
						},
					}.Ability(),
				},
			},
			OracleText: `
			At the beginning of your draw step, you may draw two additional cards. If you do, choose two cards in your hand drawn this turn. For each of those cards, pay 4 life or put the card on top of your library.
		`,
		},
	}
}
