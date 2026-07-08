package o

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
)

// ObscuraStorefront is the card definition for Obscura Storefront.
//
// Type: Land
//
// Oracle text:
//
//	When this land enters, sacrifice it. When you do, search your library for a basic Plains, Island, or Swamp card, put it onto the battlefield tapped, then shuffle and you gain 1 life.
var ObscuraStorefront = newObscuraStorefront

func newObscuraStorefront() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name:  "Obscura Storefront",
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
										Filter:       game.Selection{Supertypes: []types.Super{types.Basic}, SubtypesAny: []types.Sub{types.Sub("Plains"), types.Sub("Island"), types.Sub("Swamp")}},
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
			When this land enters, sacrifice it. When you do, search your library for a basic Plains, Island, or Swamp card, put it onto the battlefield tapped, then shuffle and you gain 1 life.
		`,
		},
	}
}
