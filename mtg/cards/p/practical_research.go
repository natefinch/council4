package p

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// PracticalResearch is the card definition for Practical Research.
//
// Type: Instant
// Cost: {3}{U}{R}
//
// Oracle text:
//
//	Draw four cards. Then discard two cards unless you discard an instant or sorcery card.
var PracticalResearch = newPracticalResearch()

func newPracticalResearch() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue, color.Red),
		CardFace: game.CardFace{
			Name: "Practical Research",
			ManaCost: opt.Val(cost.Mana{
				cost.O(3),
				cost.U,
				cost.R,
			}),
			Colors: []color.Color{color.Red, color.Blue},
			Types:  []types.Card{types.Instant},
			SpellAbility: opt.Val(game.Mode{
				Text: "Draw four cards. Then discard two cards unless you discard an instant or sorcery card.",
				Sequence: []game.Instruction{
					{
						Primitive: game.Draw{
							Amount: game.Fixed(4),
							Player: game.ControllerReference(),
						},
					},
					{
						Primitive: game.DiscardUnlessType{
							Player:      game.ControllerReference(),
							Amount:      2,
							ExemptTypes: []types.Card{types.Instant, types.Sorcery},
						},
					},
				},
			}.Ability()),
			OracleText: `
			Draw four cards. Then discard two cards unless you discard an instant or sorcery card.
		`,
		},
	}
}
