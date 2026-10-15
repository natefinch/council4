package h

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// HeroSHeirloom is the card definition for Hero's Heirloom.
//
// Type: Artifact — Equipment
// Cost: {2}
//
// Oracle text:
//
//	Equipped creature gets +2/+1.
//	As long as equipped creature is legendary, it has trample and haste.
//	Equip {2}
var HeroSHeirloom = newHeroSHeirloom

func newHeroSHeirloom() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name: "Hero's Heirloom",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
			}),
			Types:    []types.Card{types.Artifact},
			Subtypes: []types.Sub{types.Equipment},
			StaticAbilities: []game.StaticAbility{
				game.StaticAbility{
					Text: "Equipped creature gets +2/+1.",
					ContinuousEffects: []game.ContinuousEffect{
						game.ContinuousEffect{
							Duration:       game.DurationPermanent,
							Group:          game.AttachedObjectGroup(game.SourcePermanentReference()),
							Layer:          game.LayerPowerToughnessModify,
							PowerDelta:     2,
							ToughnessDelta: 1,
						},
					},
				},
				game.StaticAbility{
					Text: "As long as equipped creature is legendary, it has trample and haste.",
					Condition: opt.Val(game.Condition{
						Text:              "As long as equipped creature is legendary",
						Object:            opt.Val(game.SourceAttachedPermanentReference()),
						ObjectMatches:     opt.Val(game.Selection{Supertypes: []types.Super{types.Legendary}}),
						SourceCounterKind: counter.PlusOnePlusOne,
					}),
					ContinuousEffects: []game.ContinuousEffect{
						game.ContinuousEffect{
							Duration:    game.DurationPermanent,
							Group:       game.AttachedObjectGroup(game.SourcePermanentReference()),
							Layer:       game.LayerAbility,
							AddKeywords: []game.Keyword{game.Trample, game.Haste},
						},
					},
				},
			},
			ActivatedAbilities: []game.ActivatedAbility{
				game.EquipActivatedAbility(cost.Mana{cost.O(2)}),
			},
			OracleText: `
			Equipped creature gets +2/+1.
			As long as equipped creature is legendary, it has trample and haste.
			Equip {2}
		`,
		},
	}
}
