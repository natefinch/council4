package b

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// Batterskull is the card definition for Batterskull.
//
// Type: Artifact — Equipment
// Cost: {5}
//
// Oracle text:
//
//	Living weapon (When this Equipment enters, create a 0/0 black Phyrexian Germ creature token, then attach this to it.)
//	Equipped creature gets +4/+4 and has vigilance and lifelink.
//	{3}: Return this Equipment to its owner's hand.
//	Equip {5}
var Batterskull = newBatterskull

func newBatterskull() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name: "Batterskull",
			ManaCost: opt.Val(cost.Mana{
				cost.O(5),
			}),
			Types:    []types.Card{types.Artifact},
			Subtypes: []types.Sub{types.Equipment},
			StaticAbilities: []game.StaticAbility{
				game.StaticAbility{
					Text: "Equipped creature gets +4/+4 and has vigilance and lifelink.",
					ContinuousEffects: []game.ContinuousEffect{
						game.ContinuousEffect{
							Duration:       game.DurationPermanent,
							Group:          game.AttachedObjectGroup(game.SourcePermanentReference()),
							Layer:          game.LayerPowerToughnessModify,
							PowerDelta:     4,
							ToughnessDelta: 4,
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
				game.ActivatedAbility{
					Text: "{3}: Return this Equipment to its owner's hand.",
					ManaCost: opt.Val(cost.Mana{
						cost.O(3),
					}),
					ZoneOfFunction: zone.Battlefield,
					Content: game.AbilityContent{
						Modes: []game.Mode{
							game.Mode{
								Sequence: []game.Instruction{
									game.Instruction{
										Primitive: game.MovePermanent{
											Object:      game.SourcePermanentReference(),
											Destination: zone.Hand,
										},
									},
								},
							},
						},
						MinModes: 1,
						MaxModes: 1,
					},
				},
				game.EquipActivatedAbility(cost.Mana{cost.O(5)}),
			},
			TriggeredAbilities: []game.TriggeredAbility{
				game.LivingWeaponTriggeredAbility(),
			},
			OracleText: `
			Living weapon (When this Equipment enters, create a 0/0 black Phyrexian Germ creature token, then attach this to it.)
			Equipped creature gets +4/+4 and has vigilance and lifelink.
			{3}: Return this Equipment to its owner's hand.
			Equip {5}
		`,
		},
	}
}
