package b

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// Brainstone is the card definition for Brainstone.
//
// Type: Artifact
// Cost: {1}
//
// Oracle text:
//
//	{2}, {T}, Sacrifice this artifact: Draw three cards, then put two cards from your hand on top of your library in any order.
var Brainstone = newBrainstone

func newBrainstone() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name: "Brainstone",
			ManaCost: opt.Val(cost.Mana{
				cost.O(1),
			}),
			Types: []types.Card{types.Artifact},
			ActivatedAbilities: []game.ActivatedAbility{
				game.ActivatedAbility{
					Text: "{2}, {T}, Sacrifice this artifact: Draw three cards, then put two cards from your hand on top of your library in any order.",
					ManaCost: opt.Val(cost.Mana{
						cost.O(2),
					}),
					AdditionalCosts: []cost.Additional{
						cost.Additional{
							Kind:        cost.AdditionalTap,
							CounterKind: counter.PlusOnePlusOne,
						},
						cost.Additional{
							Kind:        cost.AdditionalSacrificeSource,
							Text:        "Sacrifice this artifact",
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
										Primitive: game.Draw{
											Amount: game.Fixed(3),
											Player: game.ControllerReference(),
										},
									},
									game.Instruction{
										Primitive: game.MoveCard{
											Player:      game.ControllerReference(),
											Amount:      game.Fixed(2),
											FromZone:    zone.Hand,
											Destination: zone.Library,
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
			{2}, {T}, Sacrifice this artifact: Draw three cards, then put two cards from your hand on top of your library in any order.
		`,
		},
	}
}
