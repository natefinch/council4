package d

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// DarksteelPlate is the card definition for Darksteel Plate.
//
// Type: Artifact — Equipment
// Cost: {3}
//
// Oracle text:
//
//	Indestructible
//	Equipped creature has indestructible.
//	Equip {2}
var DarksteelPlate = newDarksteelPlate

func newDarksteelPlate() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name: "Darksteel Plate",
			ManaCost: opt.Val(cost.Mana{
				cost.O(3),
			}),
			Types:    []types.Card{types.Artifact},
			Subtypes: []types.Sub{types.Equipment},
			StaticAbilities: []game.StaticAbility{
				game.IndestructibleStaticBody,
				game.StaticAbility{
					Text: "Equipped creature has indestructible.",
					ContinuousEffects: []game.ContinuousEffect{
						game.ContinuousEffect{
							Duration:    game.DurationPermanent,
							Group:       game.AttachedObjectGroup(game.SourcePermanentReference()),
							Layer:       game.LayerAbility,
							AddKeywords: []game.Keyword{game.Indestructible},
						},
					},
				},
			},
			ActivatedAbilities: []game.ActivatedAbility{
				game.EquipActivatedAbility(cost.Mana{cost.O(2)}),
			},
			OracleText: `
			Indestructible
			Equipped creature has indestructible.
			Equip {2}
		`,
		},
	}
}
