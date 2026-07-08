package c

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
)

// CabarettiCourtyard is the card definition for Cabaretti Courtyard.
//
// Type: Land
//
// Oracle text:
//
//	When this land enters, sacrifice it. When you do, search your library for a basic Mountain, Forest, or Plains card, put it onto the battlefield tapped, then shuffle and you gain 1 life.
var CabarettiCourtyard = newCabarettiCourtyard

func newCabarettiCourtyard() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name:  "Cabaretti Courtyard",
			Types: []types.Card{types.Land},
			TriggeredAbilities: []game.TriggeredAbility{
				game.TriggeredAbility{
					Trigger: game.TriggerCondition{
						Type: game.TriggerWhen,
						Pattern: game.TriggerPattern{
							Event:  game.EventPermanentEnteredBattlefield,
							Source: game.TriggerSourceSelf,
						},
					},
					Content: game.Mode{
						Sequence: []game.Instruction{
							{
								Primitive: game.Sacrifice{
									Object: game.EventPermanentReference(),
								},
							},
							{
								Primitive: game.Search{
									Player: game.ControllerReference(),
									Spec: game.SearchSpec{
										SourceZone:   zone.Library,
										Destination:  zone.Battlefield,
										Filter:       game.Selection{Supertypes: []types.Super{types.Basic}, SubtypesAny: []types.Sub{types.Sub("Mountain"), types.Sub("Forest"), types.Sub("Plains")}},
										EntersTapped: true,
									},
									Amount: game.Fixed(1),
								},
							},
							{
								Primitive: game.GainLife{
									Amount: game.Fixed(1),
									Player: game.ControllerReference(),
								},
							},
						},
					}.Ability(),
				},
			},
			OracleText: `
			When this land enters, sacrifice it. When you do, search your library for a basic Mountain, Forest, or Plains card, put it onto the battlefield tapped, then shuffle and you gain 1 life.
		`,
		},
	}
}
