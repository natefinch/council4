package c

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// ClawsOfGix is the card definition for Claws of Gix.
//
// Type: Artifact
// Cost: {0}
//
// Oracle text:
//
//	{1}, Sacrifice a permanent: You gain 1 life.
var ClawsOfGix = newClawsOfGix

func newClawsOfGix() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name: "Claws of Gix",
			ManaCost: opt.Val(cost.Mana{
				cost.O(0),
			}),
			Types: []types.Card{types.Artifact},
			ActivatedAbilities: []game.ActivatedAbility{
				game.ActivatedAbility{
					Text: "{1}, Sacrifice a permanent: You gain 1 life.",
					ManaCost: opt.Val(cost.Mana{
						cost.O(1),
					}),
					AdditionalCosts: []cost.Additional{
						cost.Additional{
							Kind:        cost.AdditionalSacrifice,
							Text:        "Sacrifice a permanent",
							Amount:      1,
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
											Amount: game.Fixed(1),
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
			{1}, Sacrifice a permanent: You gain 1 life.
		`,
		},
	}
}
