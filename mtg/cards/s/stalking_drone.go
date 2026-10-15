package s

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// StalkingDrone is the card definition for Stalking Drone.
//
// Type: Creature — Eldrazi Drone
// Cost: {1}{G}
//
// Oracle text:
//
//	Devoid (This card has no color.)
//	{C}: This creature gets +1/+2 until end of turn. Activate only once each turn. ({C} represents colorless mana.)
var StalkingDrone = newStalkingDrone

func newStalkingDrone() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Green),
		CardFace: game.CardFace{
			Name: "Stalking Drone",
			ManaCost: opt.Val(cost.Mana{
				cost.O(1),
				cost.G,
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
					Text: "{C}: This creature gets +1/+2 until end of turn. Activate only once each turn. ({C} represents colorless mana.)",
					ManaCost: opt.Val(cost.Mana{
						cost.C,
					}),
					ZoneOfFunction: zone.Battlefield,
					Timing:         game.OncePerTurn,
					Content: game.AbilityContent{
						Modes: []game.Mode{
							game.Mode{
								Sequence: []game.Instruction{
									game.Instruction{
										Primitive: game.ModifyPT{
											Object:         game.SourcePermanentReference(),
											PowerDelta:     game.Fixed(1),
											ToughnessDelta: game.Fixed(2),
											Duration:       game.DurationUntilEndOfTurn,
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
			{C}: This creature gets +1/+2 until end of turn. Activate only once each turn. ({C} represents colorless mana.)
		`,
		},
	}
}
