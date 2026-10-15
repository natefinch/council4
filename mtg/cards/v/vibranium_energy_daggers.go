package v

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// VibraniumEnergyDaggers is the card definition for Vibranium Energy Daggers.
//
// Type: Artifact — Equipment
// Cost: {1}
//
// Oracle text:
//
//	Indestructible (Effects that say "destroy" don't destroy this Equipment.)
//	Equipped creature gets +2/+2.
//	Equip {3} ({3}: Attach to target creature you control. Equip only as a sorcery.)
var VibraniumEnergyDaggers = newVibraniumEnergyDaggers

func newVibraniumEnergyDaggers() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name: "Vibranium Energy Daggers",
			ManaCost: opt.Val(cost.Mana{
				cost.O(1),
			}),
			Types:    []types.Card{types.Artifact},
			Subtypes: []types.Sub{types.Equipment},
			StaticAbilities: []game.StaticAbility{
				game.IndestructibleStaticBody,
				game.StaticAbility{
					Text: "Equipped creature gets +2/+2.",
					ContinuousEffects: []game.ContinuousEffect{
						game.ContinuousEffect{
							Duration:       game.DurationPermanent,
							Group:          game.AttachedObjectGroup(game.SourcePermanentReference()),
							Layer:          game.LayerPowerToughnessModify,
							PowerDelta:     2,
							ToughnessDelta: 2,
						},
					},
				},
			},
			ActivatedAbilities: []game.ActivatedAbility{
				game.EquipActivatedAbility(cost.Mana{cost.O(3)}),
			},
			OracleText: `
			Indestructible (Effects that say "destroy" don't destroy this Equipment.)
			Equipped creature gets +2/+2.
			Equip {3} ({3}: Attach to target creature you control. Equip only as a sorcery.)
		`,
		},
	}
}
