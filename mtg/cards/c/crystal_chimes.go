package c

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// CrystalChimes is the card definition for Crystal Chimes.
//
// Type: Artifact
// Cost: {3}
//
// Oracle text:
//
//	{3}, {T}, Sacrifice this artifact: Return all enchantment cards from your graveyard to your hand.
var CrystalChimes = newCrystalChimes

func newCrystalChimes() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name: "Crystal Chimes",
			ManaCost: opt.Val(cost.Mana{
				cost.O(3),
			}),
			Types: []types.Card{types.Artifact},
			ActivatedAbilities: []game.ActivatedAbility{
				game.ActivatedAbility{
					Text: "{3}, {T}, Sacrifice this artifact: Return all enchantment cards from your graveyard to your hand.",
					ManaCost: opt.Val(cost.Mana{
						cost.O(3),
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
										Primitive: game.MassReturnFromGraveyard{
											Player:      game.ControllerReference(),
											Selection:   game.Selection{RequiredTypes: []types.Card{types.Enchantment}, Controller: game.ControllerYou},
											Destination: zone.Hand,
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
			{3}, {T}, Sacrifice this artifact: Return all enchantment cards from your graveyard to your hand.
		`,
		},
	}
}
