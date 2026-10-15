package c

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// ChronicleOfVictory is the card definition for Chronicle of Victory.
//
// Type: Legendary Artifact
// Cost: {6}
//
// Oracle text:
//
//	As Chronicle of Victory enters, choose a creature type.
//	Creatures you control of the chosen type get +2/+2 and have first strike and trample.
//	Whenever you cast a spell of the chosen type, draw a card.
var ChronicleOfVictory = newChronicleOfVictory

func newChronicleOfVictory() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name: "Chronicle of Victory",
			ManaCost: opt.Val(cost.Mana{
				cost.O(6),
			}),
			Supertypes: []types.Super{types.Legendary},
			Types:      []types.Card{types.Artifact},
			StaticAbilities: []game.StaticAbility{
				game.StaticAbility{
					Text: "Creatures you control of the chosen type get +2/+2 and have first strike and trample.",
					ContinuousEffects: []game.ContinuousEffect{
						game.ContinuousEffect{
							Duration:       game.DurationPermanent,
							Group:          game.ObjectControlledGroup(game.SourcePermanentReference(), game.Selection{RequiredTypes: []types.Card{types.Creature}, SubtypeChoice: game.SubtypeChoiceSourceEntry}),
							Layer:          game.LayerPowerToughnessModify,
							PowerDelta:     2,
							ToughnessDelta: 2,
						},
						game.ContinuousEffect{
							Duration:    game.DurationPermanent,
							Group:       game.ObjectControlledGroup(game.SourcePermanentReference(), game.Selection{RequiredTypes: []types.Card{types.Creature}, SubtypeChoice: game.SubtypeChoiceSourceEntry}),
							Layer:       game.LayerAbility,
							AddKeywords: []game.Keyword{game.FirstStrike, game.Trample},
						},
					},
				},
			},
			TriggeredAbilities: []game.TriggeredAbility{
				game.TriggeredAbility{
					Text: "Whenever you cast a spell of the chosen type, draw a card.",
					Trigger: game.TriggerCondition{
						Type: game.TriggerWhenever,
						Pattern: game.TriggerPattern{
							Event:           game.EventSpellCast,
							Controller:      game.TriggerControllerYou,
							CardSelection:   game.Selection{SubtypeChoice: game.SubtypeChoiceSourceEntry},
							StackObjectKind: game.StackSpell,
							CounterKind:     counter.PlusOnePlusOne,
						},
					},
					Content: game.AbilityContent{
						Modes: []game.Mode{
							game.Mode{
								Sequence: []game.Instruction{
									game.Instruction{
										Primitive: game.Draw{
											Amount: game.Fixed(1),
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
			ReplacementAbilities: []game.ReplacementAbility{
				game.EntryTypeChoiceReplacement("As Chronicle of Victory enters, choose a creature type."),
			},
			OracleText: `
			As Chronicle of Victory enters, choose a creature type.
			Creatures you control of the chosen type get +2/+2 and have first strike and trample.
			Whenever you cast a spell of the chosen type, draw a card.
		`,
		},
	}
}
