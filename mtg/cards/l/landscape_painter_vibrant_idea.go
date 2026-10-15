package l

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// LandscapePainter is the card definition for Landscape Painter // Vibrant Idea.
//
// Type: Creature — Merfolk Wizard // Sorcery
// Cost: {1}{U} // {4}{U}
// Face: Vibrant Idea — Sorcery ({4}{U})
//
// Oracle text:
//
//	This creature enters prepared. (While it's prepared, you may cast a copy of its spell. Doing so unprepares it.)
var LandscapePainter = newLandscapePainter

func newLandscapePainter() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue),
		CardFace: game.CardFace{
			Name: "Landscape Painter",
			ManaCost: opt.Val(cost.Mana{
				cost.O(1),
				cost.U,
			}),
			Colors:         []color.Color{color.Blue},
			EntersPrepared: true,
			Types:          []types.Card{types.Creature},
			Subtypes:       []types.Sub{types.Merfolk, types.Wizard},
			Power:          opt.Val(game.PT{Value: 2}),
			Toughness:      opt.Val(game.PT{Value: 1}),
			OracleText: `
			This creature enters prepared. (While it's prepared, you may cast a copy of its spell. Doing so unprepares it.)
		`,
		},
		Layout: game.LayoutPrepare,
		Alternate: opt.Val(game.CardFace{
			Name: "Vibrant Idea",
			ManaCost: opt.Val(cost.Mana{
				cost.O(4),
				cost.U,
			}),
			Colors: []color.Color{color.Blue},
			Types:  []types.Card{types.Sorcery},
			SpellAbility: opt.Val(game.Mode{
				Sequence: []game.Instruction{
					{
						Primitive: game.Draw{
							Amount: game.Fixed(2),
							Player: game.ControllerReference(),
						},
					},
				},
			}.Ability()),
			OracleText: `
			Draw two cards.
		`,
		}),
	}
}
