package t

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// ThalakosMistfolk is the card definition for Thalakos Mistfolk.
//
// Type: Creature — Thalakos Illusion
// Cost: {2}{U}
//
// Oracle text:
//
//	Shadow (This creature can block or be blocked by only creatures with shadow.)
//	{U}: Put this creature on top of its owner's library.
var ThalakosMistfolk = newThalakosMistfolk

func newThalakosMistfolk() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue),
		CardFace: game.CardFace{
			Name: "Thalakos Mistfolk",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
				cost.U,
			}),
			Colors:    []color.Color{color.Blue},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Thalakos, types.Illusion},
			Power:     opt.Val(game.PT{Value: 2}),
			Toughness: opt.Val(game.PT{Value: 1}),
			StaticAbilities: []game.StaticAbility{
				game.ShadowStaticBody,
			},
			ActivatedAbilities: []game.ActivatedAbility{
				game.ActivatedAbility{
					Text: "{U}: Put this creature on top of its owner's library.",
					ManaCost: opt.Val(cost.Mana{
						cost.U,
					}),
					ZoneOfFunction: zone.Battlefield,
					Content: game.AbilityContent{
						Modes: []game.Mode{
							game.Mode{
								Sequence: []game.Instruction{
									game.Instruction{
										Primitive: game.MovePermanent{
											Object:      game.SourcePermanentReference(),
											Destination: zone.Library,
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
			Shadow (This creature can block or be blocked by only creatures with shadow.)
			{U}: Put this creature on top of its owner's library.
		`,
		},
	}
}
