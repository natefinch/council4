package h

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// Hivestone is the card definition for Hivestone.
//
// Type: Artifact
// Cost: {2}
//
// Oracle text:
//
//	Creatures you control are Slivers in addition to their other creature types.
var Hivestone = newHivestone

func newHivestone() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name: "Hivestone",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
			}),
			Types: []types.Card{types.Artifact},
			StaticAbilities: []game.StaticAbility{
				game.StaticAbility{
					Text: "Creatures you control are Slivers in addition to their other creature types.",
					ContinuousEffects: []game.ContinuousEffect{
						game.ContinuousEffect{
							Duration:    game.DurationPermanent,
							Group:       game.ObjectControlledGroup(game.SourcePermanentReference(), game.Selection{RequiredTypes: []types.Card{types.Creature}}),
							Layer:       game.LayerType,
							AddSubtypes: []types.Sub{types.Sliver},
						},
					},
				},
			},
			OracleText: `
			Creatures you control are Slivers in addition to their other creature types.
		`,
		},
	}
}
