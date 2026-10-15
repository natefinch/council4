package b

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// BoneSaw is the card definition for Bone Saw.
//
// Type: Artifact — Equipment
// Cost: {0}
//
// Oracle text:
//
//	Equipped creature gets +1/+0.
//	Equip {1} ({1}: Attach to target creature you control. Equip only as a sorcery.)
var BoneSaw = newBoneSaw

func newBoneSaw() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name: "Bone Saw",
			ManaCost: opt.Val(cost.Mana{
				cost.O(0),
			}),
			Types:    []types.Card{types.Artifact},
			Subtypes: []types.Sub{types.Equipment},
			StaticAbilities: []game.StaticAbility{
				game.StaticAbility{
					Text: "Equipped creature gets +1/+0.",
					ContinuousEffects: []game.ContinuousEffect{
						game.ContinuousEffect{
							Duration:   game.DurationPermanent,
							Group:      game.AttachedObjectGroup(game.SourcePermanentReference()),
							Layer:      game.LayerPowerToughnessModify,
							PowerDelta: 1,
						},
					},
				},
			},
			ActivatedAbilities: []game.ActivatedAbility{
				game.EquipActivatedAbility(cost.Mana{cost.O(1)}),
			},
			OracleText: `
			Equipped creature gets +1/+0.
			Equip {1} ({1}: Attach to target creature you control. Equip only as a sorcery.)
		`,
		},
	}
}
