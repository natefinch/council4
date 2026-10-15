package g

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// GlimmerOfGenius is the card definition for Glimmer of Genius.
//
// Type: Instant
// Cost: {3}{U}
//
// Oracle text:
//
//	Scry 2, then draw two cards. You get {E}{E} (two energy counters).
var GlimmerOfGenius = newGlimmerOfGenius

func newGlimmerOfGenius() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue),
		CardFace: game.CardFace{
			Name: "Glimmer of Genius",
			ManaCost: opt.Val(cost.Mana{
				cost.O(3),
				cost.U,
			}),
			Colors: []color.Color{color.Blue},
			Types:  []types.Card{types.Instant},
			SpellAbility: opt.Val(game.Mode{
				Sequence: []game.Instruction{
					{
						Primitive: game.Scry{
							Amount: game.Fixed(2),
							Player: game.ControllerReference(),
						},
					},
					{
						Primitive: game.Draw{
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
			Scry 2, then draw two cards. You get {E}{E} (two energy counters).
		`,
		},
	}
}
