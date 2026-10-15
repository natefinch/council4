package c

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// ChampionSHelm is the card definition for Champion's Helm.
//
// Type: Artifact — Equipment
// Cost: {3}
//
// Oracle text:
//
//	Equipped creature gets +2/+2.
//	As long as equipped creature is legendary, it has hexproof. (It can't be the target of spells or abilities your opponents control.)
//	Equip {1}
var ChampionSHelm = newChampionSHelm

func newChampionSHelm() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name: "Champion's Helm",
			ManaCost: opt.Val(cost.Mana{
				cost.O(3),
			}),
			Types:    []types.Card{types.Artifact},
			Subtypes: []types.Sub{types.Equipment},
			StaticAbilities: []game.StaticAbility{
				game.StaticAbility{
					Text: "Equipped creature gets +2/+2.",
					ContinuousEffects: []game.ContinuousEffect{
						game.ContinuousEffect{
							Duration:       game.DurationPermanent,
							Group:          game.AttachedObjectGroup(game.SourcePermanentReference()),
							Layer:          game.LayerPowerToughnessModify,
							PowerDelta:     2,
							ToughnessDelta: 2,
						},
					},
				},
				game.StaticAbility{
					Text: "As long as equipped creature is legendary, it has hexproof. (It can't be the target of spells or abilities your opponents control.)",
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
							AddKeywords: []game.Keyword{game.Hexproof},
						},
					},
				},
			},
			ActivatedAbilities: []game.ActivatedAbility{
				game.EquipActivatedAbility(cost.Mana{cost.O(1)}),
			},
			OracleText: `
			Equipped creature gets +2/+2.
			As long as equipped creature is legendary, it has hexproof. (It can't be the target of spells or abilities your opponents control.)
			Equip {1}
		`,
		},
	}
}
