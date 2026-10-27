// Legacy regression fixture preserved when the Oracle compiler was retired.
package fixtures

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
)

// RTSacrificeGrant is the card definition for RT Sacrifice Grant.
//
// Type: Artifact
//
// Oracle text:
//
//	Sacrifice a creature: Draw a card.
var RTSacrificeGrant = newRTSacrificeGrant

func newRTSacrificeGrant() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name:  "RT Sacrifice Grant",
			Types: []types.Card{types.Artifact},
			ActivatedAbilities: []game.ActivatedAbility{
				game.ActivatedAbility{
					Text: "Sacrifice a creature: Draw a card.",
					AdditionalCosts: []cost.Additional{
						cost.Additional{
							Kind:               cost.AdditionalSacrifice,
							Text:               "Sacrifice a creature",
							Amount:             1,
							MatchPermanentType: true,
							PermanentType:      types.Creature,
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
			Sacrifice a creature: Draw a card.
		`,
		},
	}
}
