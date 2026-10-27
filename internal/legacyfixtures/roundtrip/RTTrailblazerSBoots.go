// Legacy regression fixture preserved when the Oracle compiler was retired.
package fixtures

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
)

// RTTrailblazerSBoots is the card definition for RT Trailblazer's Boots.
//
// Type: Artifact — Equipment
//
// Oracle text:
//
//	Equipped creature has nonbasic landwalk. (It can't be blocked as long as defending player controls a nonbasic land.)
//	Equip {2}
var RTTrailblazerSBoots = newRTTrailblazerSBoots

func newRTTrailblazerSBoots() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name:     "RT Trailblazer's Boots",
			Types:    []types.Card{types.Artifact},
			Subtypes: []types.Sub{types.Equipment},
			StaticAbilities: []game.StaticAbility{
				game.StaticAbility{
					Text: "Equipped creature has nonbasic landwalk. (It can't be blocked as long as defending player controls a nonbasic land.)",
					ContinuousEffects: []game.ContinuousEffect{
						game.ContinuousEffect{
							Duration: game.DurationPermanent,
							Group:    game.AttachedObjectGroup(game.SourcePermanentReference()),
							Layer:    game.LayerAbility,
							AddAbilities: []game.Ability{
								&game.StaticAbility{
									Text: "Nonbasic landwalk",
									KeywordAbilities: []game.KeywordAbility{
										game.LandwalkKeyword{
											Nonbasic: true,
										},
									},
								},
							},
						},
					},
				},
			},
			ActivatedAbilities: []game.ActivatedAbility{
				game.EquipActivatedAbility(cost.Mana{cost.O(2)}),
			},
			OracleText: `
			Equipped creature has nonbasic landwalk. (It can't be blocked as long as defending player controls a nonbasic land.)
			Equip {2}
		`,
		},
	}
}
