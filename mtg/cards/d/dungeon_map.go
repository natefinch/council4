package d

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// DungeonMap is the card definition for Dungeon Map.
//
// Type: Artifact
// Cost: {3}
//
// Oracle text:
//
//	{T}: Add {C}.
//	{3}, {T}: Venture into the dungeon. Activate only as a sorcery. (Enter the first room or advance to the next room.)
var DungeonMap = newDungeonMap

func newDungeonMap() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name: "Dungeon Map",
			ManaCost: opt.Val(cost.Mana{
				cost.O(3),
			}),
			Types: []types.Card{types.Artifact},
			ActivatedAbilities: []game.ActivatedAbility{
				game.ActivatedAbility{
					Text: "{3}, {T}: Venture into the dungeon. Activate only as a sorcery. (Enter the first room or advance to the next room.)",
					ManaCost: opt.Val(cost.Mana{
						cost.O(3),
					}),
					AdditionalCosts: []cost.Additional{
						cost.Additional{
							Kind:        cost.AdditionalTap,
							CounterKind: counter.PlusOnePlusOne,
						},
					},
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
			ManaAbilities: []game.ManaAbility{
				game.TapManaAbility(mana.C),
			},
			OracleText: `
			{T}: Add {C}.
			{3}, {T}: Venture into the dungeon. Activate only as a sorcery. (Enter the first room or advance to the next room.)
		`,
		},
	}
}
