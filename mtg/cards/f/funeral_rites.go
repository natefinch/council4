package f

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// FuneralRites is the card definition for Funeral Rites.
//
// Type: Sorcery
// Cost: {2}{B}
//
// Oracle text:
//
//	You draw two cards, lose 2 life, then mill two cards.
var FuneralRites = newFuneralRites

func newFuneralRites() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Black),
		CardFace: game.CardFace{
			Name: "Funeral Rites",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
				cost.B,
			}),
			Colors: []color.Color{color.Black},
			Types:  []types.Card{types.Sorcery},
			SpellAbility: opt.Val(game.Mode{
				Sequence: []game.Instruction{
					{
						Primitive: game.Draw{
							Amount: game.Fixed(2),
							Player: game.ControllerReference(),
						},
					},
					{
						Primitive: game.LoseLife{
							Amount: game.Fixed(2),
							Player: game.ControllerReference(),
						},
					},
					{
						Primitive: game.MoveTopOfLibrary{
							Amount:      game.Fixed(2),
							Player:      game.ControllerReference(),
							Destination: zone.Graveyard,
						},
					},
				},
			}.Ability()),
			OracleText: `
			You draw two cards, lose 2 life, then mill two cards.
		`,
		},
	}
}
