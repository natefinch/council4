package t

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// TarrianSSoulcleaver is the card definition for Tarrian's Soulcleaver.
//
// Type: Legendary Artifact — Equipment
// Cost: {1}
//
// Oracle text:
//
//	Equipped creature has vigilance.
//	Whenever another artifact or creature is put into a graveyard from the battlefield, put a +1/+1 counter on equipped creature.
//	Equip {2}
var TarrianSSoulcleaver = newTarrianSSoulcleaver()

func newTarrianSSoulcleaver() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name: "Tarrian's Soulcleaver",
			ManaCost: opt.Val(cost.Mana{
				cost.O(1),
			}),
			Supertypes: []types.Super{types.Legendary},
			Types:      []types.Card{types.Artifact},
			Subtypes:   []types.Sub{types.Equipment},
			StaticAbilities: []game.StaticAbility{
				game.StaticAbility{
					ContinuousEffects: []game.ContinuousEffect{
						game.ContinuousEffect{
							Layer: game.LayerAbility,
							Group: game.AttachedObjectGroup(game.SourcePermanentReference()),
							AddKeywords: []game.Keyword{
								game.Vigilance,
							},
						},
					},
				},
			},
			ActivatedAbilities: []game.ActivatedAbility{
				game.EquipActivatedAbility(cost.Mana{cost.O(2)}),
			},
			TriggeredAbilities: []game.TriggeredAbility{
				game.TriggeredAbility{
					Trigger: game.TriggerCondition{
						Type: game.TriggerWhenever,
						Pattern: game.TriggerPattern{
							Event:            game.EventZoneChanged,
							ExcludeSelf:      true,
							MatchFromZone:    true,
							FromZone:         zone.Battlefield,
							MatchToZone:      true,
							ToZone:           zone.Graveyard,
							SubjectSelection: game.Selection{RequiredTypesAny: []types.Card{types.Artifact, types.Creature}},
						},
					},
					Content: game.Mode{
						Sequence: []game.Instruction{
							{
								Primitive: game.AddCounter{
									Amount:      game.Fixed(1),
									Object:      game.SourceAttachedPermanentReference(),
									CounterKind: counter.PlusOnePlusOne,
								},
							},
						},
					}.Ability(),
				},
			},
			OracleText: `
			Equipped creature has vigilance.
			Whenever another artifact or creature is put into a graveyard from the battlefield, put a +1/+1 counter on equipped creature.
			Equip {2}
		`,
		},
	}
}
