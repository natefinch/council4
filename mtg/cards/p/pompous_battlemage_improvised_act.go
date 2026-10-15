package p

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// PompousBattlemage is the card definition for Pompous Battlemage // Improvised Act.
//
// Type: Creature — Goblin Sorcerer // Sorcery
// Cost: {R} // {R}
// Face: Improvised Act — Sorcery ({R})
//
// Oracle text:
//
//	Prowess
//	This creature enters prepared. (While it's prepared, you may cast a copy of its spell. Doing so unprepares it.)
var PompousBattlemage = newPompousBattlemage

func newPompousBattlemage() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Red),
		CardFace: game.CardFace{
			Name: "Pompous Battlemage",
			ManaCost: opt.Val(cost.Mana{
				cost.R,
			}),
			Colors:         []color.Color{color.Red},
			EntersPrepared: true,
			Types:          []types.Card{types.Creature},
			Subtypes:       []types.Sub{types.Goblin, types.Sorcerer},
			Power:          opt.Val(game.PT{Value: 1}),
			Toughness:      opt.Val(game.PT{Value: 1}),
			StaticAbilities: []game.StaticAbility{
				game.ProwessStaticBody,
			},
			OracleText: `
			Prowess
			This creature enters prepared. (While it's prepared, you may cast a copy of its spell. Doing so unprepares it.)
		`,
		},
		Layout: game.LayoutPrepare,
		Alternate: opt.Val(game.CardFace{
			Name: "Improvised Act",
			ManaCost: opt.Val(cost.Mana{
				cost.R,
			}),
			Colors: []color.Color{color.Red},
			Types:  []types.Card{types.Sorcery},
			SpellAbility: opt.Val(game.Mode{
				Sequence: []game.Instruction{
					{
						Primitive: game.Discard{
							Amount: game.Fixed(1),
							Player: game.ControllerReference(),
						},
						Optional:      true,
						PublishResult: game.ResultKey("if-you-do"),
						LocalProducts: game.LocalProducts{
							Results: []game.ResultKey{game.ResultKey("if-you-do")},
						},
					},
					{
						Primitive: game.Draw{
							Amount: game.Fixed(1),
							Player: game.ControllerReference(),
						},
						ResultGate: opt.Val(game.InstructionResultGate{
							Key:       "if-you-do",
							Succeeded: game.TriTrue,
						}),
					},
				},
			}.Ability()),
			OracleText: `
			You may discard a card. If you do, draw a card.
		`,
		}),
	}
}
