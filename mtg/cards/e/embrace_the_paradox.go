package e

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// EmbraceTheParadox is the card definition for Embrace the Paradox.
//
// Type: Instant
// Cost: {3}{G}{U}
//
// Oracle text:
//
//	Draw three cards. You may put a land card from your hand onto the battlefield tapped.
var EmbraceTheParadox = newEmbraceTheParadox

func newEmbraceTheParadox() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue, color.Green),
		CardFace: game.CardFace{
			Name: "Embrace the Paradox",
			ManaCost: opt.Val(cost.Mana{
				cost.O(3),
				cost.G,
				cost.U,
			}),
			Colors: []color.Color{color.Green, color.Blue},
			Types:  []types.Card{types.Instant},
			SpellAbility: opt.Val(game.Mode{
				Sequence: []game.Instruction{
					{
						Primitive: game.Draw{
							Amount: game.Fixed(3),
							Player: game.ControllerReference(),
						},
					},
					{
						Primitive: game.ChooseFromZone{
							Player:     game.ControllerReference(),
							SourceZone: zone.Hand,
							Filter:     game.Selection{RequiredTypes: []types.Card{types.Land}},
							Quantity:   game.Fixed(1),
							Destination: game.ChooseDestination{
								Zone: zone.Battlefield,
							},
							Riders: game.ChooseRiders{
								EntersTapped: true,
							},
							Prompt: "Choose a card to put onto the battlefield",
						},
						Optional: true,
					},
				},
			}.Ability()),
			OracleText: `
			Draw three cards. You may put a land card from your hand onto the battlefield tapped.
		`,
		},
	}
}
