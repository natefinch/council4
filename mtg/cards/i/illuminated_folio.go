package i

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// IlluminatedFolio is the card definition for Illuminated Folio.
//
// Type: Artifact — Book
// Cost: {5}
//
// Oracle text:
//
//	{1}, {T}, Reveal two cards from your hand that share a color: Draw a card.
var IlluminatedFolio = newIlluminatedFolio

func newIlluminatedFolio() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name: "Illuminated Folio",
			ManaCost: opt.Val(cost.Mana{
				cost.O(5),
			}),
			Types:    []types.Card{types.Artifact},
			Subtypes: []types.Sub{types.Book},
			ActivatedAbilities: []game.ActivatedAbility{
				game.ActivatedAbility{
					Text: "{1}, {T}, Reveal two cards from your hand that share a color: Draw a card.",
					ManaCost: opt.Val(cost.Mana{
						cost.O(1),
					}),
					AdditionalCosts: []cost.Additional{
						cost.Additional{
							Kind:        cost.AdditionalTap,
							CounterKind: counter.PlusOnePlusOne,
						},
						cost.Additional{
							Kind:        cost.AdditionalReveal,
							Text:        "Reveal two cards from your hand that share a color",
							Amount:      2,
							Source:      zone.Hand,
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
			OracleText: `
			{1}, {T}, Reveal two cards from your hand that share a color: Draw a card.
		`,
		},
	}
}
