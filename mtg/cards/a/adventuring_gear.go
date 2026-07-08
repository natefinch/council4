package a

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// AdventuringGear is the card definition for Adventuring Gear.
//
// Type: Artifact — Equipment
// Cost: {1}
//
// Oracle text:
//
//	Landfall — Whenever a land you control enters, equipped creature gets +2/+2 until end of turn.
//	Equip {1} ({1}: Attach to target creature you control. Equip only as a sorcery.)
var AdventuringGear = newAdventuringGear

func newAdventuringGear() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name: "Adventuring Gear",
			ManaCost: opt.Val(cost.Mana{
				cost.O(1),
			}),
			Types:    []types.Card{types.Artifact},
			Subtypes: []types.Sub{types.Equipment},
			ActivatedAbilities: []game.ActivatedAbility{
				game.EquipActivatedAbility(cost.Mana{cost.O(1)}),
			},
			TriggeredAbilities: []game.TriggeredAbility{
				game.TriggeredAbility{
					Trigger: game.TriggerCondition{
						Type: game.TriggerWhenever,
						Pattern: game.TriggerPattern{
							Event:            game.EventPermanentEnteredBattlefield,
							Controller:       game.TriggerControllerYou,
							SubjectSelection: game.Selection{RequiredTypes: []types.Card{types.Land}},
						},
					},
					Content: game.Mode{
						Sequence: []game.Instruction{
							{
								Primitive: game.ApplyContinuous{
									ContinuousEffects: []game.ContinuousEffect{
										game.ContinuousEffect{
											Layer:          game.LayerPowerToughnessModify,
											Group:          game.AttachedObjectGroup(game.SourcePermanentReference()),
											PowerDelta:     2,
											ToughnessDelta: 2,
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
			Landfall — Whenever a land you control enters, equipped creature gets +2/+2 until end of turn.
			Equip {1} ({1}: Attach to target creature you control. Equip only as a sorcery.)
		`,
		},
	}
}
