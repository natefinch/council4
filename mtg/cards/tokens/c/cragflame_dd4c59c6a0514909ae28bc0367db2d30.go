package c

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
)

// Cragflame is the card definition for Cragflame.
//
// Type: Token Legendary Artifact — Equipment
//
// Oracle text:
//   Equipped creature gets +1/+1 and has vigilance, trample, and haste.
//   Equip {2}

// CragflameTokendd4c59c6a0514909ae28bc0367db2d30 is the card definition for Cragflame.
var CragflameTokendd4c59c6a0514909ae28bc0367db2d30 = newCragflameTokendd4c59c6a0514909ae28bc0367db2d30()

func newCragflameTokendd4c59c6a0514909ae28bc0367db2d30() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name:       "Cragflame",
			Supertypes: []types.Super{types.Legendary},
			Types:      []types.Card{types.Artifact},
			Subtypes:   []types.Sub{types.Equipment},
			StaticAbilities: []game.StaticAbility{
				game.StaticAbility{
					Text: "Equipped creature gets +1/+1 and has vigilance, trample, and haste.",
					ContinuousEffects: []game.ContinuousEffect{
						game.ContinuousEffect{
							Duration:       game.DurationPermanent,
							Group:          game.AttachedObjectGroup(game.SourcePermanentReference()),
							Layer:          game.LayerPowerToughnessModify,
							PowerDelta:     1,
							ToughnessDelta: 1,
						},
						game.ContinuousEffect{
							Duration:    game.DurationPermanent,
							Group:       game.AttachedObjectGroup(game.SourcePermanentReference()),
							Layer:       game.LayerAbility,
							AddKeywords: []game.Keyword{game.Vigilance, game.Trample, game.Haste},
						},
					},
				},
			},
			ActivatedAbilities: []game.ActivatedAbility{
				game.EquipActivatedAbility(cost.Mana{cost.O(2)}),
			},
			OracleText: `
			Equipped creature gets +1/+1 and has vigilance, trample, and haste.
			Equip {2}
		`,
		},
	}
}
