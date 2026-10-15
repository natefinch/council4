package a

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// AmassTheComponents is the card definition for Amass the Components.
//
// Type: Sorcery
// Cost: {3}{U}
//
// Oracle text:
//
//	Draw three cards, then put a card from your hand on the bottom of your library.
var AmassTheComponents = newAmassTheComponents

func newAmassTheComponents() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue),
		CardFace: game.CardFace{
			Name: "Amass the Components",
			ManaCost: opt.Val(cost.Mana{
				cost.O(3),
				cost.U,
			}),
			Colors: []color.Color{color.Blue},
			Types:  []types.Card{types.Sorcery},
			SpellAbility: opt.Val(game.Mode{
				Sequence: []game.Instruction{
					{
						Primitive: game.Draw{
							Amount: game.Fixed(3),
							Player: game.ControllerReference(),
						},
					},
					{
						Primitive: game.MoveCard{
							Player:            game.ControllerReference(),
							Amount:            game.Fixed(1),
							FromZone:          zone.Hand,
							Destination:       zone.Library,
							DestinationBottom: true,
						},
					},
				},
			}.Ability()),
			OracleText: `
			Draw three cards, then put a card from your hand on the bottom of your library.
		`,
		},
	}
}
