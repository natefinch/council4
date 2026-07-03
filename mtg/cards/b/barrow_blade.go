package b

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// BarrowBlade is the card definition for Barrow-Blade.
//
// Type: Artifact — Equipment
// Cost: {1}
//
// Oracle text:
//
//	Equipped creature gets +1/+1.
//	Whenever equipped creature blocks or becomes blocked by a creature, that creature loses all abilities until end of turn.
//	Equip {1} ({1}: Attach to target creature you control. Equip only as a sorcery.)
var BarrowBlade = newBarrowBlade()

func newBarrowBlade() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name: "Barrow-Blade",
			ManaCost: opt.Val(cost.Mana{
				cost.O(1),
			}),
			Types:    []types.Card{types.Artifact},
			Subtypes: []types.Sub{types.Equipment},
			StaticAbilities: []game.StaticAbility{
				game.StaticAbility{
					ContinuousEffects: []game.ContinuousEffect{
						game.ContinuousEffect{
							Layer:          game.LayerPowerToughnessModify,
							Group:          game.AttachedObjectGroup(game.SourcePermanentReference()),
							PowerDelta:     1,
							ToughnessDelta: 1,
						},
					},
				},
			},
			ActivatedAbilities: []game.ActivatedAbility{
				game.EquipActivatedAbility(cost.Mana{cost.O(1)}),
			},
			TriggeredAbilities: []game.TriggeredAbility{
				game.TriggeredAbility{
					Trigger: game.TriggerCondition{
						Type: game.TriggerWhenever,
						Pattern: game.TriggerPattern{
							Event:                   game.EventBlockerDeclared,
							Source:                  game.TriggerSourceAttachedPermanent,
							UnionEvent:              game.EventAttackerBecameBlocked,
							SubjectSelection:        game.Selection{RequiredTypes: []types.Card{types.Creature}},
							RelatedSubjectSelection: game.Selection{RequiredTypes: []types.Card{types.Creature}},
						},
					},
					Content: game.Mode{
						Sequence: []game.Instruction{
							{
								Primitive: game.ApplyContinuous{
									Object: opt.Val(game.EventRelatedPermanentReference()),
									ContinuousEffects: []game.ContinuousEffect{
										game.ContinuousEffect{
											Layer:              game.LayerAbility,
											RemoveAllAbilities: true,
										},
									},
									Duration: game.DurationUntilEndOfTurn,
								},
							},
						},
					}.Ability(),
				},
			},
			OracleText: `
			Equipped creature gets +1/+1.
			Whenever equipped creature blocks or becomes blocked by a creature, that creature loses all abilities until end of turn.
			Equip {1} ({1}: Attach to target creature you control. Equip only as a sorcery.)
		`,
		},
	}
}
