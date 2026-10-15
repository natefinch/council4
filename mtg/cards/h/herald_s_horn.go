package h

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// HeraldSHorn is the card definition for Herald's Horn.
//
// Type: Artifact
// Cost: {3}
//
// Oracle text:
//
//	As this artifact enters, choose a creature type.
//	Creature spells you cast of the chosen type cost {1} less to cast.
//	At the beginning of your upkeep, look at the top card of your library. If it's a creature card of the chosen type, you may reveal it and put it into your hand.
var HeraldSHorn = newHeraldSHorn

func newHeraldSHorn() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name: "Herald's Horn",
			ManaCost: opt.Val(cost.Mana{
				cost.O(3),
			}),
			Types: []types.Card{types.Artifact},
			StaticAbilities: []game.StaticAbility{
				game.StaticAbility{
					Text: "Creature spells you cast of the chosen type cost {1} less to cast.",
					RuleEffects: []game.RuleEffect{
						game.RuleEffect{
							Kind:           game.RuleEffectCostModifier,
							Duration:       game.DurationPermanent,
							AffectedPlayer: game.PlayerYou,
							CostModifier: game.CostModifier{
								Kind:                         game.CostModifierSpell,
								GenericReduction:             1,
								ChosenSubtypeFromEntryChoice: true,
								CardSelection:                game.Selection{RequiredTypes: []types.Card{types.Creature}},
							},
						},
					},
				},
			},
			TriggeredAbilities: []game.TriggeredAbility{
				game.TriggeredAbility{
					Text: "At the beginning of your upkeep, look at the top card of your library. If it's a creature card of the chosen type, you may reveal it and put it into your hand.",
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
								Text: "At the beginning of your upkeep, look at the top card of your library. If it's a creature card of the chosen type, you may reveal it and put it into your hand.",
								Sequence: []game.Instruction{
									game.Instruction{
										Primitive: game.LookAtLibraryTop{
											Player:        game.ControllerReference(),
											PublishLinked: game.LinkedKey("chosen-type-top"),
										},
									},
									game.Instruction{
										Primitive: game.Reveal{
											Card: game.CardReference{
												Kind:   game.CardReferenceLinked,
												LinkID: "chosen-type-top",
											},
										},
										CardCondition: opt.Val(game.CardSelection{
											Card: game.CardReference{
												Kind:   game.CardReferenceLinked,
												LinkID: "chosen-type-top",
											},
											Selection: game.Selection{RequiredTypes: []types.Card{types.Creature}, ChosenSubtypeFrom: game.EntryTypeChoiceKey},
										}),
										Optional:      true,
										PublishResult: game.ResultKey("chosen-type-revealed"),
									},
									game.Instruction{
										Primitive: game.MoveCard{
											Card: game.CardReference{
												Kind:   game.CardReferenceLinked,
												LinkID: "chosen-type-top",
											},
											FromZone:    zone.Library,
											Destination: zone.Hand,
										},
										ResultGate: opt.Val(game.InstructionResultGate{
											Key:       game.ResultKey("chosen-type-revealed"),
											Accepted:  game.TriTrue,
											Succeeded: game.TriTrue,
										}),
									},
								},
							},
						},
						MinModes: 1,
						MaxModes: 1,
					},
				},
			},
			ReplacementAbilities: []game.ReplacementAbility{
				game.EntryTypeChoiceReplacement("As this artifact enters, choose a creature type."),
			},
			OracleText: `
			As this artifact enters, choose a creature type.
			Creature spells you cast of the chosen type cost {1} less to cast.
			At the beginning of your upkeep, look at the top card of your library. If it's a creature card of the chosen type, you may reveal it and put it into your hand.
		`,
		},
	}
}
