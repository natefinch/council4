package s

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// SharpenedPitchfork is the card definition for Sharpened Pitchfork.
//
// Type: Artifact — Equipment
// Cost: {2}
//
// Oracle text:
//
//	Equipped creature has first strike.
//	As long as equipped creature is a Human, it gets +1/+1.
//	Equip {1}
var SharpenedPitchfork = newSharpenedPitchfork

func newSharpenedPitchfork() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name: "Sharpened Pitchfork",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
			}),
			Types:    []types.Card{types.Artifact},
			Subtypes: []types.Sub{types.Equipment},
			StaticAbilities: []game.StaticAbility{
				game.StaticAbility{
					ContinuousEffects: []game.ContinuousEffect{
						game.ContinuousEffect{
							Layer: game.LayerAbility,
							Group: game.AttachedObjectGroup(game.SourcePermanentReference()),
							AddKeywords: []game.Keyword{
								game.FirstStrike,
							},
						},
					},
				},
				game.StaticAbility{
					Condition: opt.Val(game.Condition{
						Object:        opt.Val(game.SourceAttachedPermanentReference()),
						ObjectMatches: opt.Val(game.Selection{SubtypesAny: []types.Sub{types.Sub("Human")}}),
					}),
					ContinuousEffects: []game.ContinuousEffect{
						game.ContinuousEffect{
							Layer:          game.LayerPowerToughnessModify,
							Group:          game.AttachedObjectGroup(game.SourcePermanentReference()),
							PowerDelta:     1,
							ToughnessDelta: 1,
						},
					},
				},
			},
			ActivatedAbilities: []game.ActivatedAbility{
				game.EquipActivatedAbility(cost.Mana{cost.O(1)}),
			},
			OracleText: `
			Equipped creature has first strike.
			As long as equipped creature is a Human, it gets +1/+1.
			Equip {1}
		`,
		},
	}
}
