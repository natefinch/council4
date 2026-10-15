package w

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// WellOfLife is the card definition for Well of Life.
//
// Type: Artifact
// Cost: {4}
//
// Oracle text:
//
//	At the beginning of your end step, if you control no untapped lands, you gain 2 life.
var WellOfLife = newWellOfLife

func newWellOfLife() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name: "Well of Life",
			ManaCost: opt.Val(cost.Mana{
				cost.O(4),
			}),
			Types: []types.Card{types.Artifact},
			TriggeredAbilities: []game.TriggeredAbility{
				game.TriggeredAbility{
					Text: "At the beginning of your end step, if you control no untapped lands, you gain 2 life.",
					Trigger: game.TriggerCondition{
						Type: game.TriggerAt,
						Pattern: game.TriggerPattern{
							Event:           game.EventBeginningOfStep,
							Controller:      game.TriggerControllerYou,
							StackObjectKind: game.StackSpell,
							CounterKind:     counter.PlusOnePlusOne,
							Step:            game.StepEnd,
						},
						InterveningIf: "if you control no untapped lands",
						InterveningCondition: opt.Val(game.Condition{
							Text:   "if you control no untapped lands",
							Negate: true,
							ControlsMatching: opt.Val(game.SelectionCount{
								Selection: game.Selection{RequiredTypes: []types.Card{types.Land}, Tapped: game.TriFalse},
								MinCount:  1,
							}),
							SourceCounterKind: counter.PlusOnePlusOne,
						}),
					},
					Content: game.AbilityContent{
						Modes: []game.Mode{
							game.Mode{
								Sequence: []game.Instruction{
									game.Instruction{
										Primitive: game.GainLife{
											Amount: game.Fixed(2),
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
			At the beginning of your end step, if you control no untapped lands, you gain 2 life.
		`,
		},
	}
}
