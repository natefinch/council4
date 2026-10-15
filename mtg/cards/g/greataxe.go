package g

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// Greataxe is the card definition for Greataxe.
//
// Type: Artifact — Equipment
// Cost: {1}
//
// Oracle text:
//
//	Equipped creature gets +4/+0.
//	Equip {5} ({5}: Attach to target creature you control. Equip only as a sorcery.)
var Greataxe = newGreataxe

func newGreataxe() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name: "Greataxe",
			ManaCost: opt.Val(cost.Mana{
				cost.O(1),
			}),
			Types:    []types.Card{types.Artifact},
			Subtypes: []types.Sub{types.Equipment},
			StaticAbilities: []game.StaticAbility{
				game.StaticAbility{
					Text: "Equipped creature gets +4/+0.",
					ContinuousEffects: []game.ContinuousEffect{
						game.ContinuousEffect{
							Duration:   game.DurationPermanent,
							Group:      game.AttachedObjectGroup(game.SourcePermanentReference()),
							Layer:      game.LayerPowerToughnessModify,
							PowerDelta: 4,
						},
					},
				},
			},
			ActivatedAbilities: []game.ActivatedAbility{
				game.EquipActivatedAbility(cost.Mana{cost.O(5)}),
			},
			OracleText: `
			Equipped creature gets +4/+0.
			Equip {5} ({5}: Attach to target creature you control. Equip only as a sorcery.)
		`,
		},
	}
}
