package m

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
)

// MaestrosTheater is the card definition for Maestros Theater.
//
// Type: Land
//
// Oracle text:
//
//	When this land enters, sacrifice it. When you do, search your library for a basic Island, Swamp, or Mountain card, put it onto the battlefield tapped, then shuffle and you gain 1 life.
var MaestrosTheater = newMaestrosTheater

func newMaestrosTheater() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name:  "Maestros Theater",
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
										Filter:       game.Selection{Supertypes: []types.Super{types.Basic}, SubtypesAny: []types.Sub{types.Sub("Island"), types.Sub("Swamp"), types.Sub("Mountain")}},
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
			When this land enters, sacrifice it. When you do, search your library for a basic Island, Swamp, or Mountain card, put it onto the battlefield tapped, then shuffle and you gain 1 life.
		`,
		},
	}
}
