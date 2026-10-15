package k

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// KiteShield is the card definition for Kite Shield.
//
// Type: Artifact — Equipment
// Cost: {0}
//
// Oracle text:
//
//	Equipped creature gets +0/+3.
//	Equip {3} ({3}: Attach to target creature you control. Equip only as a sorcery.)
var KiteShield = newKiteShield

func newKiteShield() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name: "Kite Shield",
			ManaCost: opt.Val(cost.Mana{
				cost.O(0),
			}),
			Types:    []types.Card{types.Artifact},
			Subtypes: []types.Sub{types.Equipment},
			StaticAbilities: []game.StaticAbility{
				game.StaticAbility{
					Text: "Equipped creature gets +0/+3.",
					ContinuousEffects: []game.ContinuousEffect{
						game.ContinuousEffect{
							Duration:       game.DurationPermanent,
							Group:          game.AttachedObjectGroup(game.SourcePermanentReference()),
							Layer:          game.LayerPowerToughnessModify,
							ToughnessDelta: 3,
						},
					},
				},
			},
			ActivatedAbilities: []game.ActivatedAbility{
				game.EquipActivatedAbility(cost.Mana{cost.O(3)}),
			},
			OracleText: `
			Equipped creature gets +0/+3.
			Equip {3} ({3}: Attach to target creature you control. Equip only as a sorcery.)
		`,
		},
	}
}
