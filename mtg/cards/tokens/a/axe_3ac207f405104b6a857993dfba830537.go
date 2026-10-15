package a

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
)

// Axe is the card definition for Axe.
//
// Type: Token Artifact — Equipment
//
// Oracle text:
//   Equipped creature gets +1/+0.
//   Equip {2} ({2}: Attach to target creatre you control. Equip only as a sorcery.)

// AxeToken3ac207f405104b6a857993dfba830537 is the card definition for Axe.
var AxeToken3ac207f405104b6a857993dfba830537 = newAxeToken3ac207f405104b6a857993dfba830537()

func newAxeToken3ac207f405104b6a857993dfba830537() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name:     "Axe",
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
				game.EquipActivatedAbility(cost.Mana{cost.O(2)}),
			},
			OracleText: `
			Equipped creature gets +1/+0.
			Equip {2} ({2}: Attach to target creatre you control. Equip only as a sorcery.)
		`,
		},
	}
}
