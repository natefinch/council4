package p

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// PoreOverThePages is the card definition for Pore Over the Pages.
//
// Type: Sorcery
// Cost: {3}{U}{U}
//
// Oracle text:
//
//	Draw three cards, untap up to two lands, then discard a card.
var PoreOverThePages = newPoreOverThePages

func newPoreOverThePages() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue),
		CardFace: game.CardFace{
			Name: "Pore Over the Pages",
			ManaCost: opt.Val(cost.Mana{
				cost.O(3),
				cost.U,
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
						Primitive: game.Untap{
							Group:      game.BattlefieldGroup(game.Selection{RequiredTypes: []types.Card{types.Land}}),
							ChooseUpTo: true,
							Amount:     game.Fixed(2),
						},
					},
					{
						Primitive: game.Discard{
							Amount: game.Fixed(1),
							Player: game.ControllerReference(),
						},
					},
				},
			}.Ability()),
			OracleText: `
			Draw three cards, untap up to two lands, then discard a card.
		`,
		},
	}
}
