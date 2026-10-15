package t

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// TrustyMachete is the card definition for Trusty Machete.
//
// Type: Artifact — Equipment
// Cost: {1}
//
// Oracle text:
//
//	Equipped creature gets +2/+1.
//	Equip {2}
var TrustyMachete = newTrustyMachete

func newTrustyMachete() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name: "Trusty Machete",
			ManaCost: opt.Val(cost.Mana{
				cost.O(1),
			}),
			Types:    []types.Card{types.Artifact},
			Subtypes: []types.Sub{types.Equipment},
			StaticAbilities: []game.StaticAbility{
				game.StaticAbility{
					Text: "Equipped creature gets +2/+1.",
					ContinuousEffects: []game.ContinuousEffect{
						game.ContinuousEffect{
							Duration:       game.DurationPermanent,
							Group:          game.AttachedObjectGroup(game.SourcePermanentReference()),
							Layer:          game.LayerPowerToughnessModify,
							PowerDelta:     2,
							ToughnessDelta: 1,
						},
					},
				},
			},
			ActivatedAbilities: []game.ActivatedAbility{
				game.EquipActivatedAbility(cost.Mana{cost.O(2)}),
			},
			OracleText: `
			Equipped creature gets +2/+1.
			Equip {2}
		`,
		},
	}
}
