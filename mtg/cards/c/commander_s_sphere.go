package c

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// CommanderSSphere is the card definition for Commander's Sphere.
//
// Type: Artifact
// Cost: {3}
//
// Oracle text:
//
//	{T}: Add one mana of any color in your commander's color identity.
//	Sacrifice this artifact: Draw a card.
var CommanderSSphere = newCommanderSSphere

func newCommanderSSphere() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name: "Commander's Sphere",
			ManaCost: opt.Val(cost.Mana{
				cost.O(3),
			}),
			Types: []types.Card{types.Artifact},
			ActivatedAbilities: []game.ActivatedAbility{
				game.ActivatedAbility{
					Text: "Sacrifice this artifact: Draw a card.",
					AdditionalCosts: []cost.Additional{
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
			ManaAbilities: []game.ManaAbility{
				game.TapManaCommanderIdentityAbility(),
			},
			OracleText: `
			{T}: Add one mana of any color in your commander's color identity.
			Sacrifice this artifact: Draw a card.
		`,
		},
	}
}
