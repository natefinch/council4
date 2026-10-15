package t

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// TowerOfEons is the card definition for Tower of Eons.
//
// Type: Artifact
// Cost: {4}
//
// Oracle text:
//
//	{8}, {T}: You gain 10 life.
var TowerOfEons = newTowerOfEons

func newTowerOfEons() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name: "Tower of Eons",
			ManaCost: opt.Val(cost.Mana{
				cost.O(4),
			}),
			Types: []types.Card{types.Artifact},
			ActivatedAbilities: []game.ActivatedAbility{
				game.ActivatedAbility{
					Text: "{8}, {T}: You gain 10 life.",
					ManaCost: opt.Val(cost.Mana{
						cost.O(8),
					}),
					AdditionalCosts: []cost.Additional{
						cost.Additional{
							Kind:        cost.AdditionalTap,
							CounterKind: counter.PlusOnePlusOne,
						},
					},
					ZoneOfFunction: zone.Battlefield,
					Content: game.AbilityContent{
						Modes: []game.Mode{
							game.Mode{
								Sequence: []game.Instruction{
									game.Instruction{
										Primitive: game.GainLife{
											Amount: game.Fixed(10),
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
			{8}, {T}: You gain 10 life.
		`,
		},
	}
}
