package s

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// SkithiryxTheBlightDragon is the card definition for Skithiryx, the Blight Dragon.
//
// Type: Legendary Creature — Phyrexian Dragon Skeleton
// Cost: {3}{B}{B}
//
// Oracle text:
//
//	Flying
//	Infect (This creature deals damage to creatures in the form of -1/-1 counters and to players in the form of poison counters.)
//	{B}: Skithiryx gains haste until end of turn.
//	{B}{B}: Regenerate Skithiryx.
var SkithiryxTheBlightDragon = newSkithiryxTheBlightDragon

func newSkithiryxTheBlightDragon() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Black),
		CardFace: game.CardFace{
			Name: "Skithiryx, the Blight Dragon",
			ManaCost: opt.Val(cost.Mana{
				cost.O(3),
				cost.B,
				cost.B,
			}),
			Colors:     []color.Color{color.Black},
			Supertypes: []types.Super{types.Legendary},
			Types:      []types.Card{types.Creature},
			Subtypes:   []types.Sub{types.Phyrexian, types.Dragon, types.Skeleton},
			Power:      opt.Val(game.PT{Value: 4}),
			Toughness:  opt.Val(game.PT{Value: 4}),
			StaticAbilities: []game.StaticAbility{
				game.FlyingStaticBody,
				game.InfectStaticBody,
			},
			ActivatedAbilities: []game.ActivatedAbility{
				game.ActivatedAbility{
					Text: "{B}: Skithiryx gains haste until end of turn.",
					ManaCost: opt.Val(cost.Mana{
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
													AddKeywords: []game.Keyword{game.Haste},
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
				game.ActivatedAbility{
					Text: "{B}{B}: Regenerate Skithiryx.",
					ManaCost: opt.Val(cost.Mana{
						cost.B,
						cost.B,
					}),
					ZoneOfFunction: zone.Battlefield,
					Content: game.AbilityContent{
						Modes: []game.Mode{
							game.Mode{
								Sequence: []game.Instruction{
									game.Instruction{
										Primitive: game.Regenerate{
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
			Flying
			Infect (This creature deals damage to creatures in the form of -1/-1 counters and to players in the form of poison counters.)
			{B}: Skithiryx gains haste until end of turn.
			{B}{B}: Regenerate Skithiryx.
		`,
		},
	}
}
