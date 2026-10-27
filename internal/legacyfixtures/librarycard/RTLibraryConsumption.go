// Legacy regression fixture preserved when the Oracle compiler was retired.
package fixtures

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// RTLibraryConsumption is the card definition for RT Library Consumption.
//
// Type: Sorcery
//
// Oracle text:
//
//	Reveal the top card of your library. Scry 1. If it's a creature card, put it onto the battlefield tapped under your control. Otherwise, put it into your hand.
var RTLibraryConsumption = newRTLibraryConsumption

func newRTLibraryConsumption() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name:  "RT Library Consumption",
			Types: []types.Card{types.Sorcery},
			SpellAbility: opt.Val(game.Mode{
				Sequence: []game.Instruction{
					{
						Primitive: game.Reveal{
							Amount:        game.Fixed(1),
							Player:        game.ControllerReference(),
							PublishLinked: game.LinkedKey("sequence-effect-0-product"),
						},
						LocalProducts: game.LocalProducts{
							Links: []game.LinkedKey{game.LinkedKey("sequence-effect-0-product")},
						},
					},
					{
						Primitive: game.Scry{
							Amount: game.Fixed(1),
							Player: game.ControllerReference(),
						},
					},
					{
						Primitive: game.PutOnBattlefield{
							Source:      game.LinkedBattlefieldSource(game.LinkedKey("sequence-effect-0-product")),
							Recipient:   opt.Val(game.ControllerReference()),
							EntryTapped: true,
						},
						Condition: opt.Val(game.EffectCondition{
							Condition: opt.Val(game.Condition{
								Object:        opt.Val(game.LinkedObjectReference("sequence-effect-0-product")),
								ObjectMatches: opt.Val(game.Selection{RequiredTypes: []types.Card{types.Creature}}),
							}),
						}),
						PublishCondition: "condition-0",
					},
					{
						Primitive: game.MoveCard{
							Card: game.CardReference{
								Kind:   game.CardReferenceLinked,
								LinkID: "sequence-effect-0-product",
							},
							FromZone:    zone.Library,
							Destination: zone.Hand,
						},
						ConditionGate:       "condition-0",
						ConditionGateNegate: true,
					},
				},
			}.Ability()),
			OracleText: `
			Reveal the top card of your library. Scry 1. If it's a creature card, put it onto the battlefield tapped under your control. Otherwise, put it into your hand.
		`,
		},
	}
}
