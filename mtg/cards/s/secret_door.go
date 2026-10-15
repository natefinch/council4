package s

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// SecretDoor is the card definition for Secret Door.
//
// Type: Artifact Creature — Wall
// Cost: {U}
//
// Oracle text:
//
//	Defender
//	{4}{U}: Venture into the dungeon. Activate only as a sorcery. (Enter the first room or advance to the next room.)
var SecretDoor = newSecretDoor

func newSecretDoor() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue),
		CardFace: game.CardFace{
			Name: "Secret Door",
			ManaCost: opt.Val(cost.Mana{
				cost.U,
			}),
			Colors:    []color.Color{color.Blue},
			Types:     []types.Card{types.Artifact, types.Creature},
			Subtypes:  []types.Sub{types.Wall},
			Power:     opt.Val(game.PT{Value: 0}),
			Toughness: opt.Val(game.PT{Value: 4}),
			StaticAbilities: []game.StaticAbility{
				game.DefenderStaticBody,
			},
			ActivatedAbilities: []game.ActivatedAbility{
				game.ActivatedAbility{
					Text: "{4}{U}: Venture into the dungeon. Activate only as a sorcery. (Enter the first room or advance to the next room.)",
					ManaCost: opt.Val(cost.Mana{
						cost.O(4),
						cost.U,
					}),
					ZoneOfFunction: zone.Battlefield,
					Timing:         game.SorceryOnly,
					Content: game.AbilityContent{
						Modes: []game.Mode{
							game.Mode{
								Sequence: []game.Instruction{
									game.Instruction{
										Primitive: game.VentureIntoDungeon{
											Player: game.ControllerReference(),
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
			Defender
			{4}{U}: Venture into the dungeon. Activate only as a sorcery. (Enter the first room or advance to the next room.)
		`,
		},
	}
}
