package t

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// ThranQuarry is the card definition for Thran Quarry.
//
// Type: Land
//
// Oracle text:
//
//	At the beginning of the end step, if you control no creatures, sacrifice this land.
//	{T}: Add one mana of any color.
var ThranQuarry = newThranQuarry

func newThranQuarry() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name:  "Thran Quarry",
			Types: []types.Card{types.Land},
			ManaAbilities: []game.ManaAbility{
				game.TapManaChoiceAbility(mana.W, mana.U, mana.B, mana.R, mana.G),
			},
			TriggeredAbilities: []game.TriggeredAbility{
				game.TriggeredAbility{
					Text: "At the beginning of the end step, if you control no creatures, sacrifice this land.",
					Trigger: game.TriggerCondition{
						Type: game.TriggerAt,
						Pattern: game.TriggerPattern{
							Event:           game.EventBeginningOfStep,
							StackObjectKind: game.StackSpell,
							CounterKind:     counter.PlusOnePlusOne,
							Step:            game.StepEnd,
						},
						InterveningIf: "if you control no creatures",
						InterveningCondition: opt.Val(game.Condition{
							Text:   "if you control no creatures",
							Negate: true,
							ControlsMatching: opt.Val(game.SelectionCount{
								Selection: game.Selection{RequiredTypes: []types.Card{types.Creature}},
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
										Primitive: game.Sacrifice{
											Object: game.SourcePermanentReference(),
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
			At the beginning of the end step, if you control no creatures, sacrifice this land.
			{T}: Add one mana of any color.
		`,
		},
	}
}
