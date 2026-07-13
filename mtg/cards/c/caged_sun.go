package c

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// CagedSun is the card definition for Caged Sun.
//
// Type: Artifact
// Cost: {6}
//
// Oracle text:
//
//	As this artifact enters, choose a color.
//	Creatures you control of the chosen color get +1/+1.
//	Whenever a land's ability causes you to add one or more mana of the chosen color, add an additional one mana of that color.
var CagedSun = newCagedSun

func newCagedSun() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name: "Caged Sun",
			ManaCost: opt.Val(cost.Mana{
				cost.O(6),
			}),
			Types: []types.Card{types.Artifact},
			StaticAbilities: []game.StaticAbility{
				game.StaticAbility{
					ContinuousEffects: []game.ContinuousEffect{
						game.ContinuousEffect{
							Layer:          game.LayerPowerToughnessModify,
							Group:          game.ObjectControlledGroup(game.SourcePermanentReference(), game.Selection{RequiredTypes: []types.Card{types.Creature}, ColorChoice: game.ColorChoiceSourceEntry}),
							PowerDelta:     1,
							ToughnessDelta: 1,
						},
					},
				},
			},
			TriggeredAbilities: []game.TriggeredAbility{
				game.TriggeredAbility{
					Trigger: game.TriggerCondition{
						Type: game.TriggerWhenever,
						Pattern: game.TriggerPattern{
							Event:                                   game.EventPermanentTapped,
							Controller:                              game.TriggerControllerYou,
							RequireTappedForMana:                    true,
							RequireProducedManaColorFromEntryChoice: true,
							SubjectSelection:                        game.Selection{RequiredTypes: []types.Card{types.Land}},
						},
					},
					Content: game.Mode{
						Sequence: []game.Instruction{
							{
								Primitive: game.AddMana{
									Amount:          game.Fixed(1),
									EntryChoiceFrom: game.ChoiceKey("oracle-entry-color"),
								},
							},
						},
					}.Ability(),
				},
			},
			ReplacementAbilities: []game.ReplacementAbility{
				game.EntryColorChoiceReplacement("As this artifact enters, choose a color."),
			},
			OracleText: `
			As this artifact enters, choose a color.
			Creatures you control of the chosen color get +1/+1.
			Whenever a land's ability causes you to add one or more mana of the chosen color, add an additional one mana of that color.
		`,
		},
	}
}
