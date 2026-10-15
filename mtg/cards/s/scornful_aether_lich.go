package s

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// ScornfulAetherLich is the card definition for Scornful Aether-Lich.
//
// Type: Artifact Creature — Zombie Wizard
// Cost: {3}{U}
//
// Oracle text:
//
//	{W}{B}: This creature gains fear and vigilance until end of turn. (Attacking doesn't cause it to tap, and it can't be blocked except by artifact creatures and/or black creatures.)
var ScornfulAetherLich = newScornfulAetherLich

func newScornfulAetherLich() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White, color.Blue, color.Black),
		CardFace: game.CardFace{
			Name: "Scornful Aether-Lich",
			ManaCost: opt.Val(cost.Mana{
				cost.O(3),
				cost.U,
			}),
			Colors:    []color.Color{color.Blue},
			Types:     []types.Card{types.Artifact, types.Creature},
			Subtypes:  []types.Sub{types.Zombie, types.Wizard},
			Power:     opt.Val(game.PT{Value: 2}),
			Toughness: opt.Val(game.PT{Value: 4}),
			ActivatedAbilities: []game.ActivatedAbility{
				game.ActivatedAbility{
					Text: "{W}{B}: This creature gains fear and vigilance until end of turn. (Attacking doesn't cause it to tap, and it can't be blocked except by artifact creatures and/or black creatures.)",
					ManaCost: opt.Val(cost.Mana{
						cost.W,
						cost.B,
					}),
					ZoneOfFunction: zone.Battlefield,
					Content: game.AbilityContent{
						Modes: []game.Mode{
							game.Mode{
								Sequence: []game.Instruction{
									game.Instruction{
										Primitive: game.ApplyContinuous{
											Object: opt.Val(game.SourcePermanentReference()),
											ContinuousEffects: []game.ContinuousEffect{
												game.ContinuousEffect{
													Duration:    game.DurationPermanent,
													Layer:       game.LayerAbility,
													AddKeywords: []game.Keyword{game.Fear, game.Vigilance},
												},
											},
											Duration: game.DurationUntilEndOfTurn,
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
			{W}{B}: This creature gains fear and vigilance until end of turn. (Attacking doesn't cause it to tap, and it can't be blocked except by artifact creatures and/or black creatures.)
		`,
		},
	}
}
