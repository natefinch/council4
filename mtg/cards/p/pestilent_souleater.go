package p

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// PestilentSouleater is the card definition for Pestilent Souleater.
//
// Type: Artifact Creature — Phyrexian Insect
// Cost: {5}
//
// Oracle text:
//
//	{B/P}: This creature gains infect until end of turn. ({B/P} can be paid with either {B} or 2 life. A creature with infect deals damage to creatures in the form of -1/-1 counters and to players in the form of poison counters.)
var PestilentSouleater = newPestilentSouleater

func newPestilentSouleater() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Black),
		CardFace: game.CardFace{
			Name: "Pestilent Souleater",
			ManaCost: opt.Val(cost.Mana{
				cost.O(5),
			}),
			Types:     []types.Card{types.Artifact, types.Creature},
			Subtypes:  []types.Sub{types.Phyrexian, types.Insect},
			Power:     opt.Val(game.PT{Value: 3}),
			Toughness: opt.Val(game.PT{Value: 3}),
			ActivatedAbilities: []game.ActivatedAbility{
				game.ActivatedAbility{
					Text: "{B/P}: This creature gains infect until end of turn. ({B/P} can be paid with either {B} or 2 life. A creature with infect deals damage to creatures in the form of -1/-1 counters and to players in the form of poison counters.)",
					ManaCost: opt.Val(cost.Mana{
						cost.PhyrexianMana(mana.B),
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
													AddKeywords: []game.Keyword{game.Infect},
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
			{B/P}: This creature gains infect until end of turn. ({B/P} can be paid with either {B} or 2 life. A creature with infect deals damage to creatures in the form of -1/-1 counters and to players in the form of poison counters.)
		`,
		},
	}
}
