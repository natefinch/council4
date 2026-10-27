// Legacy regression fixture preserved when the Oracle compiler was retired.
package fixtures

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
)

// RTBrotherhoodRegalia is the card definition for RT Brotherhood Regalia.
//
// Type: Artifact — Equipment
//
// Oracle text:
//
//	Equipped creature has ward {2}, is an Assassin in addition to its other types, and can't be blocked.
//	Equip legendary creature {1}
//	Equip {3}
var RTBrotherhoodRegalia = newRTBrotherhoodRegalia

func newRTBrotherhoodRegalia() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name:     "RT Brotherhood Regalia",
			Types:    []types.Card{types.Artifact},
			Subtypes: []types.Sub{types.Equipment},
			StaticAbilities: []game.StaticAbility{
				game.StaticAbility{
					Text: "Equipped creature has ward {2}, is an Assassin in addition to its other types, and can't be blocked.",
					ContinuousEffects: []game.ContinuousEffect{
						game.ContinuousEffect{
							Duration: game.DurationPermanent,
							Group:    game.AttachedObjectGroup(game.SourcePermanentReference()),
							Layer:    game.LayerAbility,
							AddAbilities: []game.Ability{
								new(game.WardStaticAbility(cost.Mana{cost.O(2)})),
							},
						},
						game.ContinuousEffect{
							Duration:    game.DurationPermanent,
							Group:       game.AttachedObjectGroup(game.SourcePermanentReference()),
							Layer:       game.LayerType,
							AddSubtypes: []types.Sub{types.Assassin},
						},
					},
					RuleEffects: []game.RuleEffect{
						game.RuleEffect{
							Kind:             game.RuleEffectCantBeBlocked,
							AffectedAttached: true,
							Duration:         game.DurationPermanent,
						},
					},
				},
			},
			ActivatedAbilities: []game.ActivatedAbility{
				game.EquipRestrictedActivatedAbility(cost.Mana{cost.O(1)}, []types.Super{types.Legendary}, nil),
				game.EquipActivatedAbility(cost.Mana{cost.O(3)}),
			},
			OracleText: `
			Equipped creature has ward {2}, is an Assassin in addition to its other types, and can't be blocked.
			Equip legendary creature {1}
			Equip {3}
		`,
		},
	}
}
