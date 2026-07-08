package d

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// DoorOfDestinies is the card definition for Door of Destinies.
//
// Type: Artifact
// Cost: {4}
//
// Oracle text:
//
//	As this artifact enters, choose a creature type.
//	Whenever you cast a spell of the chosen type, put a charge counter on this artifact.
//	Creatures you control of the chosen type get +1/+1 for each charge counter on this artifact.
var DoorOfDestinies = newDoorOfDestinies

func newDoorOfDestinies() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name: "Door of Destinies",
			ManaCost: opt.Val(cost.Mana{
				cost.O(4),
			}),
			Types: []types.Card{types.Artifact},
			StaticAbilities: []game.StaticAbility{
				game.StaticAbility{
					ContinuousEffects: []game.ContinuousEffect{
						game.ContinuousEffect{
							Layer: game.LayerPowerToughnessModify,
							Group: game.ObjectControlledGroup(game.SourcePermanentReference(), game.Selection{RequiredTypes: []types.Card{types.Creature}, SubtypeChoice: game.SubtypeChoiceSourceEntry}),
							PowerDeltaDynamic: opt.Val(game.DynamicAmount{
								Kind:        game.DynamicAmountObjectCounters,
								Multiplier:  1,
								CounterKind: counter.Charge,
								Object:      game.SourcePermanentReference(),
							}),
							ToughnessDeltaDynamic: opt.Val(game.DynamicAmount{
								Kind:        game.DynamicAmountObjectCounters,
								Multiplier:  1,
								CounterKind: counter.Charge,
								Object:      game.SourcePermanentReference(),
							}),
						},
					},
				},
			},
			TriggeredAbilities: []game.TriggeredAbility{
				game.TriggeredAbility{
					Trigger: game.TriggerCondition{
						Type: game.TriggerWhenever,
						Pattern: game.TriggerPattern{
							Event:         game.EventSpellCast,
							Controller:    game.TriggerControllerYou,
							CardSelection: game.Selection{SubtypeChoice: game.SubtypeChoiceSourceEntry},
						},
					},
					Content: game.Mode{
						Sequence: []game.Instruction{
							{
								Primitive: game.AddCounter{
									Amount:      game.Fixed(1),
									Object:      game.SourcePermanentReference(),
									CounterKind: counter.Charge,
								},
							},
						},
					}.Ability(),
				},
			},
			ReplacementAbilities: []game.ReplacementAbility{
				game.EntryTypeChoiceReplacement("As this artifact enters, choose a creature type."),
			},
			OracleText: `
			As this artifact enters, choose a creature type.
			Whenever you cast a spell of the chosen type, put a charge counter on this artifact.
			Creatures you control of the chosen type get +1/+1 for each charge counter on this artifact.
		`,
		},
	}
}
