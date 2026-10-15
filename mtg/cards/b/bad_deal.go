package b

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// BadDeal is the card definition for Bad Deal.
//
// Type: Sorcery
// Cost: {4}{B}{B}
//
// Oracle text:
//
//	You draw two cards and each opponent discards two cards. Each player loses 2 life.
var BadDeal = newBadDeal

func newBadDeal() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Black),
		CardFace: game.CardFace{
			Name: "Bad Deal",
			ManaCost: opt.Val(cost.Mana{
				cost.O(4),
				cost.B,
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
						Primitive: game.Discard{
							Amount:      game.Fixed(2),
							PlayerGroup: game.OpponentsReference(),
						},
					},
					{
						Primitive: game.LoseLife{
							Amount:      game.Fixed(2),
							PlayerGroup: game.AllPlayersReference(),
						},
					},
				},
			}.Ability()),
			OracleText: `
			You draw two cards and each opponent discards two cards. Each player loses 2 life.
		`,
		},
	}
}
