package h

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// HowlingMine is the card definition for Howling Mine.
//
// Type: Artifact
// Cost: {2}
//
// Oracle text:
//
//	At the beginning of each player's draw step, if this artifact is untapped, that player draws an additional card.
var HowlingMine = newHowlingMine

func newHowlingMine() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name: "Howling Mine",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
			}),
			Types: []types.Card{types.Artifact},
			TriggeredAbilities: []game.TriggeredAbility{
				game.TriggeredAbility{
					Text: "At the beginning of each player's draw step, if this artifact is untapped, that player draws an additional card.",
					Trigger: game.TriggerCondition{
						Type: game.TriggerAt,
						Pattern: game.TriggerPattern{
							Event:           game.EventBeginningOfStep,
							StackObjectKind: game.StackSpell,
							CounterKind:     counter.PlusOnePlusOne,
							Step:            game.StepDraw,
						},
						InterveningIf: "if this artifact is untapped",
						InterveningCondition: opt.Val(game.Condition{
							Text:              "if this artifact is untapped",
							Object:            opt.Val(game.SourcePermanentReference()),
							ObjectMatches:     opt.Val(game.Selection{RequiredTypes: []types.Card{types.Artifact}, Tapped: game.TriFalse}),
							SourceCounterKind: counter.PlusOnePlusOne,
						}),
					},
					Content: game.AbilityContent{
						Modes: []game.Mode{
							game.Mode{
								Sequence: []game.Instruction{
									game.Instruction{
										Primitive: game.Draw{
											Amount: game.Fixed(1),
											Player: game.EventPlayerReference(),
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
			At the beginning of each player's draw step, if this artifact is untapped, that player draws an additional card.
		`,
		},
	}
}
