package a

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// AmorphousAxe is the card definition for Amorphous Axe.
//
// Type: Artifact — Equipment
// Cost: {2}
//
// Oracle text:
//
//	Equipped creature gets +3/+0 and is every creature type.
//	Equip {3} ({3}: Attach to target creature you control. Equip only as a sorcery.)
var AmorphousAxe = newAmorphousAxe

func newAmorphousAxe() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name: "Amorphous Axe",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
			}),
			Types:    []types.Card{types.Artifact},
			Subtypes: []types.Sub{types.Equipment},
			StaticAbilities: []game.StaticAbility{
				game.StaticAbility{
					Text: "Equipped creature gets +3/+0 and is every creature type.",
					ContinuousEffects: []game.ContinuousEffect{
						game.ContinuousEffect{
							Duration:   game.DurationPermanent,
							Group:      game.AttachedObjectGroup(game.SourcePermanentReference()),
							Layer:      game.LayerPowerToughnessModify,
							PowerDelta: 3,
						},
						game.ContinuousEffect{
							Duration:             game.DurationPermanent,
							Group:                game.AttachedObjectGroup(game.SourcePermanentReference()),
							Layer:                game.LayerType,
							AddEveryCreatureType: true,
						},
					},
				},
			},
			ActivatedAbilities: []game.ActivatedAbility{
				game.EquipActivatedAbility(cost.Mana{cost.O(3)}),
			},
			OracleText: `
			Equipped creature gets +3/+0 and is every creature type.
			Equip {3} ({3}: Attach to target creature you control. Equip only as a sorcery.)
		`,
		},
	}
}
