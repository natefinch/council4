package b

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// BladedBracers is the card definition for Bladed Bracers.
//
// Type: Artifact — Equipment
// Cost: {1}
//
// Oracle text:
//
//	Equipped creature gets +1/+1.
//	As long as equipped creature is a Human or an Angel, it has vigilance.
//	Equip {2} ({2}: Attach to target creature you control. Equip only as a sorcery.)
var BladedBracers = newBladedBracers()

func newBladedBracers() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name: "Bladed Bracers",
			ManaCost: opt.Val(cost.Mana{
				cost.O(1),
			}),
			Types:    []types.Card{types.Artifact},
			Subtypes: []types.Sub{types.Equipment},
			StaticAbilities: []game.StaticAbility{
				game.StaticAbility{
					ContinuousEffects: []game.ContinuousEffect{
						game.ContinuousEffect{
							Layer:          game.LayerPowerToughnessModify,
							Group:          game.AttachedObjectGroup(game.SourcePermanentReference()),
							PowerDelta:     1,
							ToughnessDelta: 1,
						},
					},
				},
				game.StaticAbility{
					Condition: opt.Val(game.Condition{
						Object:        opt.Val(game.SourceAttachedPermanentReference()),
						ObjectMatches: opt.Val(game.Selection{SubtypesAny: []types.Sub{types.Sub("Human"), types.Sub("Angel")}}),
					}),
					ContinuousEffects: []game.ContinuousEffect{
						game.ContinuousEffect{
							Layer: game.LayerAbility,
							Group: game.AttachedObjectGroup(game.SourcePermanentReference()),
							AddKeywords: []game.Keyword{
								game.Vigilance,
							},
						},
					},
				},
			},
			ActivatedAbilities: []game.ActivatedAbility{
				game.EquipActivatedAbility(cost.Mana{cost.O(2)}),
			},
			OracleText: `
			Equipped creature gets +1/+1.
			As long as equipped creature is a Human or an Angel, it has vigilance.
			Equip {2} ({2}: Attach to target creature you control. Equip only as a sorcery.)
		`,
		},
	}
}
