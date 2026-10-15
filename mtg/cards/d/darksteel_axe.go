package d

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// DarksteelAxe is the card definition for Darksteel Axe.
//
// Type: Artifact — Equipment
// Cost: {1}
//
// Oracle text:
//
//	Indestructible (Effects that say "destroy" don't destroy this Equipment.)
//	Equipped creature gets +2/+0.
//	Equip {2}
var DarksteelAxe = newDarksteelAxe

func newDarksteelAxe() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name: "Darksteel Axe",
			ManaCost: opt.Val(cost.Mana{
				cost.O(1),
			}),
			Types:    []types.Card{types.Artifact},
			Subtypes: []types.Sub{types.Equipment},
			StaticAbilities: []game.StaticAbility{
				game.IndestructibleStaticBody,
				game.StaticAbility{
					Text: "Equipped creature gets +2/+0.",
					ContinuousEffects: []game.ContinuousEffect{
						game.ContinuousEffect{
							Duration:   game.DurationPermanent,
							Group:      game.AttachedObjectGroup(game.SourcePermanentReference()),
							Layer:      game.LayerPowerToughnessModify,
							PowerDelta: 2,
						},
					},
				},
			},
			ActivatedAbilities: []game.ActivatedAbility{
				game.EquipActivatedAbility(cost.Mana{cost.O(2)}),
			},
			OracleText: `
			Indestructible (Effects that say "destroy" don't destroy this Equipment.)
			Equipped creature gets +2/+0.
			Equip {2}
		`,
		},
	}
}
