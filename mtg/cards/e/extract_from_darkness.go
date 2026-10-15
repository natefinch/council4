package e

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// ExtractFromDarkness is the card definition for Extract from Darkness.
//
// Type: Sorcery
// Cost: {3}{U}{B}
//
// Oracle text:
//
//	Each player mills two cards. Then you put a creature card from a graveyard onto the battlefield under your control.
var ExtractFromDarkness = newExtractFromDarkness

func newExtractFromDarkness() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue, color.Black),
		CardFace: game.CardFace{
			Name: "Extract from Darkness",
			ManaCost: opt.Val(cost.Mana{
				cost.O(3),
				cost.U,
				cost.B,
			}),
			Colors: []color.Color{color.Black, color.Blue},
			Types:  []types.Card{types.Sorcery},
			SpellAbility: opt.Val(game.Mode{
				Sequence: []game.Instruction{
					{
						Primitive: game.MoveTopOfLibrary{
							Amount:      game.Fixed(2),
							PlayerGroup: game.AllPlayersReference(),
							Destination: zone.Graveyard,
						},
					},
					{
						Primitive: game.ChooseFromZone{
							Player:     game.ControllerReference(),
							SourceZone: zone.Graveyard,
							AllOwners:  true,
							Filter:     game.Selection{RequiredTypes: []types.Card{types.Creature}},
							Quantity:   game.Fixed(1),
							Destination: game.ChooseDestination{
								Zone: zone.Battlefield,
							},
							Prompt: "Choose a card to return to the battlefield",
						},
					},
				},
			}.Ability()),
			OracleText: `
			Each player mills two cards. Then you put a creature card from a graveyard onto the battlefield under your control.
		`,
		},
	}
}
