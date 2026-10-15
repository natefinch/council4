package s

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
)

// SturdyShield is the card definition for Sturdy Shield.
//
// Type: Token Artifact — Equipment
//
// Oracle text:
//   Equipped creature gets +1/+2.
//   Equip {2} ({2}: Attach to target creature you control. Equip only as a sorcery.)

// SturdyShieldTokenc949b3f6c0424323b36c7f10391c8610 is the card definition for Sturdy Shield.
var SturdyShieldTokenc949b3f6c0424323b36c7f10391c8610 = newSturdyShieldTokenc949b3f6c0424323b36c7f10391c8610()

func newSturdyShieldTokenc949b3f6c0424323b36c7f10391c8610() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name:     "Sturdy Shield",
			Types:    []types.Card{types.Artifact},
			Subtypes: []types.Sub{types.Equipment},
			StaticAbilities: []game.StaticAbility{
				game.StaticAbility{
					Text: "Equipped creature gets +1/+2.",
					ContinuousEffects: []game.ContinuousEffect{
						game.ContinuousEffect{
							Duration:       game.DurationPermanent,
							Group:          game.AttachedObjectGroup(game.SourcePermanentReference()),
							Layer:          game.LayerPowerToughnessModify,
							PowerDelta:     1,
							ToughnessDelta: 2,
						},
					},
				},
			},
			ActivatedAbilities: []game.ActivatedAbility{
				game.EquipActivatedAbility(cost.Mana{cost.O(2)}),
			},
			OracleText: `
			Equipped creature gets +1/+2.
			Equip {2} ({2}: Attach to target creature you control. Equip only as a sorcery.)
		`,
		},
	}
}
