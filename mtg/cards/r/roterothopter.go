package r

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// Roterothopter is the card definition for Roterothopter.
//
// Type: Artifact Creature — Thopter
// Cost: {1}
//
// Oracle text:
//
//	Flying
//	{2}: This creature gets +1/+0 until end of turn. Activate no more than twice each turn.
var Roterothopter = newRoterothopter

func newRoterothopter() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name: "Roterothopter",
			ManaCost: opt.Val(cost.Mana{
				cost.O(1),
			}),
			Types:     []types.Card{types.Artifact, types.Creature},
			Subtypes:  []types.Sub{types.Thopter},
			Power:     opt.Val(game.PT{Value: 0}),
			Toughness: opt.Val(game.PT{Value: 2}),
			StaticAbilities: []game.StaticAbility{
				game.FlyingStaticBody,
			},
			ActivatedAbilities: []game.ActivatedAbility{
				game.ActivatedAbility{
					Text: "{2}: This creature gets +1/+0 until end of turn. Activate no more than twice each turn.",
					ManaCost: opt.Val(cost.Mana{
						cost.O(2),
					}),
					ZoneOfFunction:        zone.Battlefield,
					MaxActivationsPerTurn: 2,
					Content: game.AbilityContent{
						Modes: []game.Mode{
							game.Mode{
								Sequence: []game.Instruction{
									game.Instruction{
										Primitive: game.ModifyPT{
											Object:     game.SourcePermanentReference(),
											PowerDelta: game.Fixed(1),
											Duration:   game.DurationUntilEndOfTurn,
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
			Flying
			{2}: This creature gets +1/+0 until end of turn. Activate no more than twice each turn.
		`,
		},
	}
}
