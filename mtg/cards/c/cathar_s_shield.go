package c

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// CatharSShield is the card definition for Cathar's Shield.
//
// Type: Artifact — Equipment
// Cost: {0}
//
// Oracle text:
//
//	Equipped creature gets +0/+3 and has vigilance.
//	Equip {3} ({3}: Attach to target creature you control. Equip only as a sorcery.)
var CatharSShield = newCatharSShield

func newCatharSShield() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name: "Cathar's Shield",
			ManaCost: opt.Val(cost.Mana{
				cost.O(0),
			}),
			Types:    []types.Card{types.Artifact},
			Subtypes: []types.Sub{types.Equipment},
			StaticAbilities: []game.StaticAbility{
				game.StaticAbility{
					Text: "Equipped creature gets +0/+3 and has vigilance.",
					ContinuousEffects: []game.ContinuousEffect{
						game.ContinuousEffect{
							Duration:       game.DurationPermanent,
							Group:          game.AttachedObjectGroup(game.SourcePermanentReference()),
							Layer:          game.LayerPowerToughnessModify,
							ToughnessDelta: 3,
						},
						game.ContinuousEffect{
							Duration:    game.DurationPermanent,
							Group:       game.AttachedObjectGroup(game.SourcePermanentReference()),
							Layer:       game.LayerAbility,
							AddKeywords: []game.Keyword{game.Vigilance},
						},
					},
				},
			},
			ActivatedAbilities: []game.ActivatedAbility{
				game.EquipActivatedAbility(cost.Mana{cost.O(3)}),
			},
			OracleText: `
			Equipped creature gets +0/+3 and has vigilance.
			Equip {3} ({3}: Attach to target creature you control. Equip only as a sorcery.)
		`,
		},
	}
}
