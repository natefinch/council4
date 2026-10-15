package s

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// SwordOfKaldra is the card definition for Sword of Kaldra.
//
// Type: Legendary Artifact — Equipment
// Cost: {4}
//
// Oracle text:
//
//	Equipped creature gets +5/+5.
//	Whenever equipped creature deals damage to a creature, exile that creature. (Exile it only if it's still on the battlefield.)
//	Equip {4} ({4}: Attach to target creature you control. Equip only as a sorcery.)
var SwordOfKaldra = newSwordOfKaldra

func newSwordOfKaldra() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name: "Sword of Kaldra",
			ManaCost: opt.Val(cost.Mana{
				cost.O(4),
			}),
			Supertypes: []types.Super{types.Legendary},
			Types:      []types.Card{types.Artifact},
			Subtypes:   []types.Sub{types.Equipment},
			StaticAbilities: []game.StaticAbility{
				game.StaticAbility{
					Text: "Equipped creature gets +5/+5.",
					ContinuousEffects: []game.ContinuousEffect{
						game.ContinuousEffect{
							Duration:       game.DurationPermanent,
							Group:          game.AttachedObjectGroup(game.SourcePermanentReference()),
							Layer:          game.LayerPowerToughnessModify,
							PowerDelta:     5,
							ToughnessDelta: 5,
						},
					},
				},
			},
			ActivatedAbilities: []game.ActivatedAbility{
				game.EquipActivatedAbility(cost.Mana{cost.O(4)}),
			},
			TriggeredAbilities: []game.TriggeredAbility{
				game.TriggeredAbility{
					Text: "Whenever equipped creature deals damage to a creature, exile that creature. (Exile it only if it's still on the battlefield.)",
					Trigger: game.TriggerCondition{
						Type: game.TriggerWhenever,
						Pattern: game.TriggerPattern{
							Event:                 game.EventDamageDealt,
							Source:                game.TriggerSourceAttachedPermanent,
							Subject:               game.TriggerSubjectDamageSource,
							StackObjectKind:       game.StackSpell,
							DamageRecipient:       game.DamageRecipientPermanent,
							DamageRecipientTypes:  []types.Card{types.Creature},
							DamageSourceSelection: game.Selection{RequiredTypes: []types.Card{types.Creature}},
							CounterKind:           counter.PlusOnePlusOne,
						},
					},
					Content: game.AbilityContent{
						Modes: []game.Mode{
							game.Mode{
								Sequence: []game.Instruction{
									game.Instruction{
										Primitive: game.MovePermanent{
											Object:      game.EventPermanentReference(),
											Destination: zone.Exile,
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
			Equipped creature gets +5/+5.
			Whenever equipped creature deals damage to a creature, exile that creature. (Exile it only if it's still on the battlefield.)
			Equip {4} ({4}: Attach to target creature you control. Equip only as a sorcery.)
		`,
		},
	}
}
