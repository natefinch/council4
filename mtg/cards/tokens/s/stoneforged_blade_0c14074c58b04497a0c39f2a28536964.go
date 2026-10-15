package s

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
)

// StoneforgedBlade is the card definition for Stoneforged Blade.
//
// Type: Token Artifact — Equipment
//
// Oracle text:
//   Indestructible
//   Equipped creature gets +5/+5 and has double strike.
//   Equip {0}

// StoneforgedBladeToken0c14074c58b04497a0c39f2a28536964 is the card definition for Stoneforged Blade.
var StoneforgedBladeToken0c14074c58b04497a0c39f2a28536964 = newStoneforgedBladeToken0c14074c58b04497a0c39f2a28536964()

func newStoneforgedBladeToken0c14074c58b04497a0c39f2a28536964() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name:     "Stoneforged Blade",
			Types:    []types.Card{types.Artifact},
			Subtypes: []types.Sub{types.Equipment},
			StaticAbilities: []game.StaticAbility{
				game.IndestructibleStaticBody,
				game.StaticAbility{
					Text: "Equipped creature gets +5/+5 and has double strike.",
					ContinuousEffects: []game.ContinuousEffect{
						game.ContinuousEffect{
							Duration:       game.DurationPermanent,
							Group:          game.AttachedObjectGroup(game.SourcePermanentReference()),
							Layer:          game.LayerPowerToughnessModify,
							PowerDelta:     5,
							ToughnessDelta: 5,
						},
						game.ContinuousEffect{
							Duration:    game.DurationPermanent,
							Group:       game.AttachedObjectGroup(game.SourcePermanentReference()),
							Layer:       game.LayerAbility,
							AddKeywords: []game.Keyword{game.DoubleStrike},
						},
					},
				},
			},
			ActivatedAbilities: []game.ActivatedAbility{
				game.EquipActivatedAbility(cost.Mana{cost.O(0)}),
			},
			OracleText: `
			Indestructible
			Equipped creature gets +5/+5 and has double strike.
			Equip {0}
		`,
		},
	}
}
