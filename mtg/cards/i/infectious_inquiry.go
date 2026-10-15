package i

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// InfectiousInquiry is the card definition for Infectious Inquiry.
//
// Type: Sorcery
// Cost: {2}{B}
//
// Oracle text:
//
//	You draw two cards and you lose 2 life. Each opponent gets a poison counter.
var InfectiousInquiry = newInfectiousInquiry

func newInfectiousInquiry() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Black),
		CardFace: game.CardFace{
			Name: "Infectious Inquiry",
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
						Primitive: game.AddPlayerCounter{
							Amount:      game.Fixed(1),
							PlayerGroup: game.OpponentsReference(),
							CounterKind: counter.Poison,
						},
					},
				},
			}.Ability()),
			OracleText: `
			You draw two cards and you lose 2 life. Each opponent gets a poison counter.
		`,
		},
	}
}
