package s

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// Skinwing is the card definition for Skinwing.
//
// Type: Artifact — Equipment
// Cost: {4}
//
// Oracle text:
//
//	Living weapon (When this Equipment enters, create a 0/0 black Phyrexian Germ creature token, then attach this to it.)
//	Equipped creature gets +2/+2 and has flying.
//	Equip {6}
var Skinwing = newSkinwing

func newSkinwing() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name: "Skinwing",
			ManaCost: opt.Val(cost.Mana{
				cost.O(4),
			}),
			Types:    []types.Card{types.Artifact},
			Subtypes: []types.Sub{types.Equipment},
			StaticAbilities: []game.StaticAbility{
				game.StaticAbility{
					Text: "Equipped creature gets +2/+2 and has flying.",
					ContinuousEffects: []game.ContinuousEffect{
						game.ContinuousEffect{
							Duration:       game.DurationPermanent,
							Group:          game.AttachedObjectGroup(game.SourcePermanentReference()),
							Layer:          game.LayerPowerToughnessModify,
							PowerDelta:     2,
							ToughnessDelta: 2,
						},
						game.ContinuousEffect{
							Duration:    game.DurationPermanent,
							Group:       game.AttachedObjectGroup(game.SourcePermanentReference()),
							Layer:       game.LayerAbility,
							AddKeywords: []game.Keyword{game.Flying},
						},
					},
				},
			},
			ActivatedAbilities: []game.ActivatedAbility{
				game.EquipActivatedAbility(cost.Mana{cost.O(6)}),
			},
			TriggeredAbilities: []game.TriggeredAbility{
				game.LivingWeaponTriggeredAbility(),
			},
			OracleText: `
			Living weapon (When this Equipment enters, create a 0/0 black Phyrexian Germ creature token, then attach this to it.)
			Equipped creature gets +2/+2 and has flying.
			Equip {6}
		`,
		},
	}
}
