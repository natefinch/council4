package s

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// SlaughterDrone is the card definition for Slaughter Drone.
//
// Type: Creature — Eldrazi Drone
// Cost: {1}{B}
//
// Oracle text:
//
//	Devoid (This card has no color.)
//	{C}: This creature gains deathtouch until end of turn. (Any amount of damage it deals to a creature is enough to destroy it. {C} represents colorless mana.)
var SlaughterDrone = newSlaughterDrone

func newSlaughterDrone() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Black),
		CardFace: game.CardFace{
			Name: "Slaughter Drone",
			ManaCost: opt.Val(cost.Mana{
				cost.O(1),
				cost.B,
			}),
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Eldrazi, types.Drone},
			Power:     opt.Val(game.PT{Value: 2}),
			Toughness: opt.Val(game.PT{Value: 2}),
			StaticAbilities: []game.StaticAbility{
				game.DevoidStaticBody,
			},
			ActivatedAbilities: []game.ActivatedAbility{
				game.ActivatedAbility{
					Text: "{C}: This creature gains deathtouch until end of turn. (Any amount of damage it deals to a creature is enough to destroy it. {C} represents colorless mana.)",
					ManaCost: opt.Val(cost.Mana{
						cost.C,
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
													AddKeywords: []game.Keyword{game.Deathtouch},
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
			Devoid (This card has no color.)
			{C}: This creature gains deathtouch until end of turn. (Any amount of damage it deals to a creature is enough to destroy it. {C} represents colorless mana.)
		`,
		},
	}
}
