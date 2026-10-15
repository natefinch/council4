package g

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// GrappleWithThePast is the card definition for Grapple with the Past.
//
// Type: Instant
// Cost: {1}{G}
//
// Oracle text:
//
//	Mill three cards, then you may return a creature or land card from your graveyard to your hand. (To mill three cards, put the top three cards of your library into your graveyard.)
var GrappleWithThePast = newGrappleWithThePast

func newGrappleWithThePast() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Green),
		CardFace: game.CardFace{
			Name: "Grapple with the Past",
			ManaCost: opt.Val(cost.Mana{
				cost.O(1),
				cost.G,
			}),
			Colors: []color.Color{color.Green},
			Types:  []types.Card{types.Instant},
			SpellAbility: opt.Val(game.Mode{
				Sequence: []game.Instruction{
					{
						Primitive: game.MoveTopOfLibrary{
							Amount:      game.Fixed(3),
							Player:      game.ControllerReference(),
							Destination: zone.Graveyard,
						},
					},
					{
						Primitive: game.ChooseFromZone{
							Player:     game.ControllerReference(),
							SourceZone: zone.Graveyard,
							Filter:     game.Selection{RequiredTypesAny: []types.Card{types.Creature, types.Land}, Controller: game.ControllerYou},
							Quantity:   game.Fixed(1),
							Destination: game.ChooseDestination{
								Zone: zone.Hand,
							},
							Prompt: "Choose a card to return to your hand",
						},
						Optional: true,
					},
				},
			}.Ability()),
			OracleText: `
			Mill three cards, then you may return a creature or land card from your graveyard to your hand. (To mill three cards, put the top three cards of your library into your graveyard.)
		`,
		},
	}
}
