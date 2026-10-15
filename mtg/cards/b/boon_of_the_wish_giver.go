package b

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// BoonOfTheWishGiver is the card definition for Boon of the Wish-Giver.
//
// Type: Sorcery
// Cost: {4}{U}{U}
//
// Oracle text:
//
//	Draw four cards.
//	Cycling {1} ({1}, Discard this card: Draw a card.)
var BoonOfTheWishGiver = newBoonOfTheWishGiver

func newBoonOfTheWishGiver() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue),
		CardFace: game.CardFace{
			Name: "Boon of the Wish-Giver",
			ManaCost: opt.Val(cost.Mana{
				cost.O(4),
				cost.U,
				cost.U,
			}),
			Colors: []color.Color{color.Blue},
			Types:  []types.Card{types.Sorcery},
			ActivatedAbilities: []game.ActivatedAbility{
				game.CyclingActivatedAbility(cost.Mana{cost.O(1)}),
			},
			SpellAbility: opt.Val(game.Mode{
				Sequence: []game.Instruction{
					{
						Primitive: game.Draw{
							Amount: game.Fixed(4),
							Player: game.ControllerReference(),
						},
					},
				},
			}.Ability()),
			OracleText: `
			Draw four cards.
			Cycling {1} ({1}, Discard this card: Draw a card.)
		`,
		},
	}
}
