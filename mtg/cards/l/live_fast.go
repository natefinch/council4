package l

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// LiveFast is the card definition for Live Fast.
//
// Type: Sorcery
// Cost: {2}{B}
//
// Oracle text:
//
//	You draw two cards, lose 2 life, and get {E}{E} (two energy counters).
var LiveFast = newLiveFast

func newLiveFast() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Black),
		CardFace: game.CardFace{
			Name: "Live Fast",
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
							Amount:      game.Fixed(2),
							Player:      game.ControllerReference(),
							CounterKind: counter.Energy,
						},
					},
				},
			}.Ability()),
			OracleText: `
			You draw two cards, lose 2 life, and get {E}{E} (two energy counters).
		`,
		},
	}
}
