package c

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// CorpseChurn is the card definition for Corpse Churn.
//
// Type: Instant
// Cost: {1}{B}
//
// Oracle text:
//
//	Mill three cards, then you may return a creature card from your graveyard to your hand. (To mill three cards, put the top three cards of your library into your graveyard.)
var CorpseChurn = newCorpseChurn

func newCorpseChurn() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Black),
		CardFace: game.CardFace{
			Name: "Corpse Churn",
			ManaCost: opt.Val(cost.Mana{
				cost.O(1),
				cost.B,
			}),
			Colors: []color.Color{color.Black},
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
							Filter:     game.Selection{RequiredTypes: []types.Card{types.Creature}, Controller: game.ControllerYou},
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
			Mill three cards, then you may return a creature card from your graveyard to your hand. (To mill three cards, put the top three cards of your library into your graveyard.)
		`,
		},
	}
}
