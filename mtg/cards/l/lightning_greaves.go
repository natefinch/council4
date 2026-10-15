package l

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// LightningGreaves is the card definition for Lightning Greaves.
//
// Type: Artifact — Equipment
// Cost: {2}
//
// Oracle text:
//
//	Equipped creature has haste and shroud. (It can't be the target of spells or abilities.)
//	Equip {0}
var LightningGreaves = newLightningGreaves

func newLightningGreaves() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name: "Lightning Greaves",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
			}),
			Types:    []types.Card{types.Artifact},
			Subtypes: []types.Sub{types.Equipment},
			StaticAbilities: []game.StaticAbility{
				game.StaticAbility{
					Text: "Equipped creature has haste and shroud. (It can't be the target of spells or abilities.)",
					ContinuousEffects: []game.ContinuousEffect{
						game.ContinuousEffect{
							Duration:    game.DurationPermanent,
							Group:       game.AttachedObjectGroup(game.SourcePermanentReference()),
							Layer:       game.LayerAbility,
							AddKeywords: []game.Keyword{game.Haste, game.Shroud},
						},
					},
				},
			},
			ActivatedAbilities: []game.ActivatedAbility{
				game.EquipActivatedAbility(cost.Mana{cost.O(0)}),
			},
			OracleText: `
			Equipped creature has haste and shroud. (It can't be the target of spells or abilities.)
			Equip {0}
		`,
		},
	}
}
