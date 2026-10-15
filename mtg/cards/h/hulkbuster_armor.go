package h

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// HulkbusterArmor is the card definition for Hulkbuster Armor.
//
// Type: Legendary Artifact — Equipment
// Cost: {4}
//
// Oracle text:
//
//	Equipped creature has base power and toughness 9/9 and has flying.
//	Equip Hero {3}
//	Equip {6}
var HulkbusterArmor = newHulkbusterArmor

func newHulkbusterArmor() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name: "Hulkbuster Armor",
			ManaCost: opt.Val(cost.Mana{
				cost.O(4),
			}),
			Supertypes: []types.Super{types.Legendary},
			Types:      []types.Card{types.Artifact},
			Subtypes:   []types.Sub{types.Equipment},
			StaticAbilities: []game.StaticAbility{
				game.StaticAbility{
					Text: "Equipped creature has base power and toughness 9/9 and has flying.",
					ContinuousEffects: []game.ContinuousEffect{
						game.ContinuousEffect{
							Duration: game.DurationPermanent,
							Group:    game.AttachedObjectGroup(game.SourcePermanentReference()),
							Layer:    game.LayerPowerToughnessSet,
							SetPower: opt.Val(game.PT{
								Value: 9,
							}),
							SetToughness: opt.Val(game.PT{
								Value: 9,
							}),
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
				game.EquipRestrictedActivatedAbility(cost.Mana{cost.O(3)}, nil, []types.Sub{types.Hero}),
				game.EquipActivatedAbility(cost.Mana{cost.O(6)}),
			},
			OracleText: `
			Equipped creature has base power and toughness 9/9 and has flying.
			Equip Hero {3}
			Equip {6}
		`,
		},
	}
}
