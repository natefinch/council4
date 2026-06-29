package t

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// ThirstForKnowledge is the card definition for Thirst for Knowledge.
//
// Type: Instant
// Cost: {2}{U}
//
// Oracle text:
//
//	Draw three cards. Then discard two cards unless you discard an artifact card.
var ThirstForKnowledge = newThirstForKnowledge()

func newThirstForKnowledge() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue),
		CardFace: game.CardFace{
			Name: "Thirst for Knowledge",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
				cost.U,
			}),
			Colors: []color.Color{color.Blue},
			Types:  []types.Card{types.Instant},
			SpellAbility: opt.Val(game.Mode{
				Text: "Draw three cards. Then discard two cards unless you discard an artifact card.",
				Sequence: []game.Instruction{
					{
						Primitive: game.Draw{
							Amount: game.Fixed(3),
							Player: game.ControllerReference(),
						},
					},
					{
						Primitive: game.DiscardUnlessType{
							Player:      game.ControllerReference(),
							Amount:      2,
							ExemptTypes: []types.Card{types.Artifact},
						},
					},
				},
			}.Ability()),
			OracleText: `
			Draw three cards. Then discard two cards unless you discard an artifact card.
		`,
		},
	}
}
