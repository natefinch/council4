package s

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// SlagwurmArmor is the card definition for Slagwurm Armor.
//
// Type: Artifact — Equipment
// Cost: {1}
//
// Oracle text:
//
//	Equipped creature gets +0/+6.
//	Equip {3} ({3}: Attach to target creature you control. Equip only as a sorcery. This card enters unattached and stays on the battlefield if the creature leaves.)
var SlagwurmArmor = newSlagwurmArmor

func newSlagwurmArmor() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name: "Slagwurm Armor",
			ManaCost: opt.Val(cost.Mana{
				cost.O(1),
			}),
			Types:    []types.Card{types.Artifact},
			Subtypes: []types.Sub{types.Equipment},
			StaticAbilities: []game.StaticAbility{
				game.StaticAbility{
					Text: "Equipped creature gets +0/+6.",
					ContinuousEffects: []game.ContinuousEffect{
						game.ContinuousEffect{
							Duration:       game.DurationPermanent,
							Group:          game.AttachedObjectGroup(game.SourcePermanentReference()),
							Layer:          game.LayerPowerToughnessModify,
							ToughnessDelta: 6,
						},
					},
				},
			},
			ActivatedAbilities: []game.ActivatedAbility{
				game.EquipActivatedAbility(cost.Mana{cost.O(3)}),
			},
			OracleText: `
			Equipped creature gets +0/+6.
			Equip {3} ({3}: Attach to target creature you control. Equip only as a sorcery. This card enters unattached and stays on the battlefield if the creature leaves.)
		`,
		},
	}
}
