package p

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// PrologueToPhyresis is the card definition for Prologue to Phyresis.
//
// Type: Instant
// Cost: {1}{U}
//
// Oracle text:
//
//	Each opponent gets a poison counter.
//	Draw a card.
var PrologueToPhyresis = newPrologueToPhyresis

func newPrologueToPhyresis() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue),
		CardFace: game.CardFace{
			Name: "Prologue to Phyresis",
			ManaCost: opt.Val(cost.Mana{
				cost.O(1),
				cost.U,
			}),
			Colors: []color.Color{color.Blue},
			Types:  []types.Card{types.Instant},
			SpellAbility: opt.Val(game.Mode{
				Sequence: []game.Instruction{
					{
						Primitive: game.AddPlayerCounter{
							Amount:      game.Fixed(1),
							PlayerGroup: game.OpponentsReference(),
							CounterKind: counter.Poison,
						},
					},
					{
						Primitive: game.Draw{
							Amount: game.Fixed(1),
							Player: game.ControllerReference(),
						},
					},
				},
			}.Ability()),
			OracleText: `
			Each opponent gets a poison counter.
			Draw a card.
		`,
		},
	}
}
