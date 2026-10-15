package c

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// ChariotOfVictory is the card definition for Chariot of Victory.
//
// Type: Artifact — Equipment
// Cost: {3}
//
// Oracle text:
//
//	Equipped creature has first strike, trample, and haste.
//	Equip {1}
var ChariotOfVictory = newChariotOfVictory

func newChariotOfVictory() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name: "Chariot of Victory",
			ManaCost: opt.Val(cost.Mana{
				cost.O(3),
			}),
			Types:    []types.Card{types.Artifact},
			Subtypes: []types.Sub{types.Equipment},
			StaticAbilities: []game.StaticAbility{
				game.StaticAbility{
					Text: "Equipped creature has first strike, trample, and haste.",
					ContinuousEffects: []game.ContinuousEffect{
						game.ContinuousEffect{
							Duration:    game.DurationPermanent,
							Group:       game.AttachedObjectGroup(game.SourcePermanentReference()),
							Layer:       game.LayerAbility,
							AddKeywords: []game.Keyword{game.FirstStrike, game.Trample, game.Haste},
						},
					},
				},
			},
			ActivatedAbilities: []game.ActivatedAbility{
				game.EquipActivatedAbility(cost.Mana{cost.O(1)}),
			},
			OracleText: `
			Equipped creature has first strike, trample, and haste.
			Equip {1}
		`,
		},
	}
}
