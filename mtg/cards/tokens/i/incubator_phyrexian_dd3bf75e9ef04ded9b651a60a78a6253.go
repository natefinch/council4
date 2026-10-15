package i

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// Incubator is the card definition for Incubator // Phyrexian.
//
// Type: Token Artifact — Incubator // Token Artifact Creature — Phyrexian
// Face: Phyrexian — Token Artifact Creature — Phyrexian
//
// Oracle text:
//   {2}: Transform this artifact.

// IncubatorTokendd3bf75e9ef04ded9b651a60a78a6253 is the card definition for Incubator.
var IncubatorTokendd3bf75e9ef04ded9b651a60a78a6253 = newIncubatorTokendd3bf75e9ef04ded9b651a60a78a6253()

func newIncubatorTokendd3bf75e9ef04ded9b651a60a78a6253() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name:     "Incubator",
			Types:    []types.Card{types.Artifact},
			Subtypes: []types.Sub{types.Incubator},
			ActivatedAbilities: []game.ActivatedAbility{
				game.ActivatedAbility{
					Text: "{2}: Transform this artifact.",
					ManaCost: opt.Val(cost.Mana{
						cost.O(2),
					}),
					ZoneOfFunction: zone.Battlefield,
					Content: game.AbilityContent{
						Modes: []game.Mode{
							game.Mode{
								Sequence: []game.Instruction{
									game.Instruction{
										Primitive: game.Transform{
											Object: game.SourcePermanentReference(),
										},
									},
								},
							},
						},
						MinModes: 1,
						MaxModes: 1,
					},
				},
			},
			OracleText: `
			{2}: Transform this artifact.
		`,
		},
		Layout: game.LayoutDoubleFacedToken,
		Back: opt.Val(game.CardFace{
			Name:      "Phyrexian",
			Types:     []types.Card{types.Artifact, types.Creature},
			Subtypes:  []types.Sub{types.Phyrexian},
			Power:     opt.Val(game.PT{Value: 0}),
			Toughness: opt.Val(game.PT{Value: 0}),
		}),
	}
}
