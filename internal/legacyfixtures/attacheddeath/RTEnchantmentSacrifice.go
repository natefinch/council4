// Legacy regression fixture preserved when the Oracle compiler was retired.
package fixtures

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
)

// RTEnchantmentSacrifice is the card definition for RT Enchantment Sacrifice.
//
// Type: Artifact
//
// Oracle text:
//
//	Sacrifice an enchantment: Draw a card.
var RTEnchantmentSacrifice = newRTEnchantmentSacrifice

func newRTEnchantmentSacrifice() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name:  "RT Enchantment Sacrifice",
			Types: []types.Card{types.Artifact},
			ActivatedAbilities: []game.ActivatedAbility{
				game.ActivatedAbility{
					Text: "Sacrifice an enchantment: Draw a card.",
					AdditionalCosts: []cost.Additional{
						cost.Additional{
							Kind:               cost.AdditionalSacrifice,
							Text:               "Sacrifice an enchantment",
							Amount:             1,
							MatchPermanentType: true,
							PermanentType:      types.Enchantment,
							CounterKind:        counter.PlusOnePlusOne,
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
			Sacrifice an enchantment: Draw a card.
		`,
		},
	}
}
