package g

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// GenjiGlove is the card definition for Genji Glove.
//
// Type: Artifact — Equipment
// Cost: {5}
//
// Oracle text:
//
//	Equipped creature has double strike.
//	Whenever equipped creature attacks, if it's the first combat phase of the turn, untap it. After this phase, there is an additional combat phase.
//	Equip {3}
var GenjiGlove = newGenjiGlove

func newGenjiGlove() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name: "Genji Glove",
			ManaCost: opt.Val(cost.Mana{
				cost.O(5),
			}),
			Types:    []types.Card{types.Artifact},
			Subtypes: []types.Sub{types.Equipment},
			StaticAbilities: []game.StaticAbility{
				game.StaticAbility{
					Text: "Equipped creature has double strike.",
					ContinuousEffects: []game.ContinuousEffect{
						game.ContinuousEffect{
							Duration:    game.DurationPermanent,
							Group:       game.AttachedObjectGroup(game.SourcePermanentReference()),
							Layer:       game.LayerAbility,
							AddKeywords: []game.Keyword{game.DoubleStrike},
						},
					},
				},
			},
			ActivatedAbilities: []game.ActivatedAbility{
				game.EquipActivatedAbility(cost.Mana{cost.O(3)}),
			},
			TriggeredAbilities: []game.TriggeredAbility{
				game.TriggeredAbility{
					Text: "Whenever equipped creature attacks, if it's the first combat phase of the turn, untap it. After this phase, there is an additional combat phase.",
					Trigger: game.TriggerCondition{
						Type: game.TriggerWhenever,
						Pattern: game.TriggerPattern{
							Event:            game.EventAttackerDeclared,
							Source:           game.TriggerSourceAttachedPermanent,
							SubjectSelection: game.Selection{RequiredTypes: []types.Card{types.Creature}},
							StackObjectKind:  game.StackSpell,
							CounterKind:      counter.PlusOnePlusOne,
						},
						InterveningIf: "if it's the first combat phase of the turn",
						InterveningCondition: opt.Val(game.Condition{
							Text:                   "if it's the first combat phase of the turn",
							SourceCounterKind:      counter.PlusOnePlusOne,
							FirstCombatPhaseOfTurn: true,
						}),
					},
					Content: game.AbilityContent{
						Modes: []game.Mode{
							game.Mode{
								Sequence: []game.Instruction{
									game.Instruction{
										Primitive: game.Untap{
											Object: game.EventPermanentReference(),
										},
									},
									game.Instruction{
										Primitive: game.AddExtraPhases{
											Combat: true,
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
			Equipped creature has double strike.
			Whenever equipped creature attacks, if it's the first combat phase of the turn, untap it. After this phase, there is an additional combat phase.
			Equip {3}
		`,
		},
	}
}
