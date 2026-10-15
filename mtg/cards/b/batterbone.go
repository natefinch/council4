package b

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// Batterbone is the card definition for Batterbone.
//
// Type: Artifact — Equipment
// Cost: {2}
//
// Oracle text:
//
//	Living weapon (When this Equipment enters, create a 0/0 black Phyrexian Germ creature token, then attach this to it.)
//	Equipped creature gets +1/+1 and has vigilance and lifelink.
//	Equip {5}
var Batterbone = newBatterbone

func newBatterbone() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name: "Batterbone",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
			}),
			Types:    []types.Card{types.Artifact},
			Subtypes: []types.Sub{types.Equipment},
			StaticAbilities: []game.StaticAbility{
				game.StaticAbility{
					Text: "Equipped creature gets +1/+1 and has vigilance and lifelink.",
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
							AddKeywords: []game.Keyword{game.Vigilance, game.Lifelink},
						},
					},
				},
			},
			ActivatedAbilities: []game.ActivatedAbility{
				game.EquipActivatedAbility(cost.Mana{cost.O(5)}),
			},
			TriggeredAbilities: []game.TriggeredAbility{
				game.LivingWeaponTriggeredAbility(),
			},
			OracleText: `
			Living weapon (When this Equipment enters, create a 0/0 black Phyrexian Germ creature token, then attach this to it.)
			Equipped creature gets +1/+1 and has vigilance and lifelink.
			Equip {5}
		`,
		},
	}
}
