package i

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// InThePresenceOfAges is the card definition for In the Presence of Ages.
//
// Type: Instant
// Cost: {2}{G}
//
// Oracle text:
//
//	Reveal the top four cards of your library. You may put a creature card and/or a land card from among them into your hand. Put the rest into your graveyard.
var InThePresenceOfAges = newInThePresenceOfAges

func newInThePresenceOfAges() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Green),
		CardFace: game.CardFace{
			Name: "In the Presence of Ages",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
				cost.G,
			}),
			Colors: []color.Color{color.Green},
			Types:  []types.Card{types.Instant},
			SpellAbility: opt.Val(game.Mode{
				Sequence: []game.Instruction{
					{
						Primitive: game.Mill{
							Amount:        game.Fixed(4),
							Player:        game.ControllerReference(),
							PublishLinked: game.LinkedKey("milled-cards"),
						},
					},
					{
						Primitive: game.ChooseFromZone{
							Player:     game.ControllerReference(),
							SourceZone: zone.Graveyard,
							Filter:     game.Selection{RequiredTypes: []types.Card{types.Creature}},
							Quantity:   game.Fixed(1),
							Destination: game.ChooseDestination{
								Zone: zone.Hand,
							},
							Riders: game.ChooseRiders{
								FromLinked: game.LinkedKey("milled-cards"),
							},
							Prompt: "Choose a card to return to your hand",
						},
						Optional: true,
					},
					{
						Primitive: game.ChooseFromZone{
							Player:     game.ControllerReference(),
							SourceZone: zone.Graveyard,
							Filter:     game.Selection{RequiredTypes: []types.Card{types.Land}},
							Quantity:   game.Fixed(1),
							Destination: game.ChooseDestination{
								Zone: zone.Hand,
							},
							Riders: game.ChooseRiders{
								FromLinked: game.LinkedKey("milled-cards"),
							},
							Prompt: "Choose a card to return to your hand",
						},
						Optional: true,
					},
				},
			}.Ability()),
			OracleText: `
			Reveal the top four cards of your library. You may put a creature card and/or a land card from among them into your hand. Put the rest into your graveyard.
		`,
		},
	}
}
