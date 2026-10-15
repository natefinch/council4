package t

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// TrailblazerSBoots is the card definition for Trailblazer's Boots.
//
// Type: Artifact — Equipment
// Cost: {2}
//
// Oracle text:
//
//	Equipped creature has nonbasic landwalk. (It can't be blocked as long as defending player controls a nonbasic land.)
//	Equip {2}
var TrailblazerSBoots = newTrailblazerSBoots

func newTrailblazerSBoots() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name: "Trailblazer's Boots",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
			}),
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
