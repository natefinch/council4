package h

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// HallOfTriumph is the card definition for Hall of Triumph.
//
// Type: Legendary Artifact
// Cost: {3}
//
// Oracle text:
//
//	As Hall of Triumph enters, choose a color.
//	Creatures you control of the chosen color get +1/+1.
var HallOfTriumph = newHallOfTriumph

func newHallOfTriumph() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name: "Hall of Triumph",
			ManaCost: opt.Val(cost.Mana{
				cost.O(3),
			}),
			Supertypes: []types.Super{types.Legendary},
			Types:      []types.Card{types.Artifact},
			StaticAbilities: []game.StaticAbility{
				game.StaticAbility{
					Text: "Creatures you control of the chosen color get +1/+1.",
					ContinuousEffects: []game.ContinuousEffect{
						game.ContinuousEffect{
							Duration:       game.DurationPermanent,
							Group:          game.ObjectControlledGroup(game.SourcePermanentReference(), game.Selection{RequiredTypes: []types.Card{types.Creature}, ColorChoice: game.ColorChoiceSourceEntry}),
							Layer:          game.LayerPowerToughnessModify,
							PowerDelta:     1,
							ToughnessDelta: 1,
						},
					},
				},
			},
			ReplacementAbilities: []game.ReplacementAbility{
				game.EntryColorChoiceReplacement("As Hall of Triumph enters, choose a color."),
			},
			OracleText: `
			As Hall of Triumph enters, choose a color.
			Creatures you control of the chosen color get +1/+1.
		`,
		},
	}
}
