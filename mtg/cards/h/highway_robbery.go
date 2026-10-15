package h

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// HighwayRobbery is the card definition for Highway Robbery.
//
// Type: Sorcery
// Cost: {1}{R}
//
// Oracle text:
//
//	You may discard a card or sacrifice a land. If you do, draw two cards.
//	Plot {1}{R} (You may pay {1}{R} and exile this card from your hand. Cast it as a sorcery on a later turn without paying its mana cost. Plot only as a sorcery.)
var HighwayRobbery = newHighwayRobbery

func newHighwayRobbery() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Red),
		CardFace: game.CardFace{
			Name: "Highway Robbery",
			ManaCost: opt.Val(cost.Mana{
				cost.O(1),
				cost.R,
			}),
			Colors: []color.Color{color.Red},
			Types:  []types.Card{types.Sorcery},
			StaticAbilities: []game.StaticAbility{
				game.StaticAbility{
					Text: "Plot {1}{R}",
					KeywordAbilities: []game.KeywordAbility{
						game.PlotKeyword{
							Cost: cost.Mana{
								cost.O(1),
								cost.R,
							},
						},
					},
				},
			},
			SpellAbility: opt.Val(game.Mode{
				Sequence: []game.Instruction{
					{
						Primitive: game.Discard{
							Amount: game.Fixed(1),
							Player: game.ControllerReference(),
						},
						Optional:      true,
						PublishResult: game.ResultKey("disjunctive-cost-a"),
					},
					{
						Primitive: game.SacrificePermanents{
							Player:    game.ControllerReference(),
							Amount:    game.Fixed(1),
							Selection: game.Selection{RequiredTypes: []types.Card{types.Land}},
						},
						ResultGate: opt.Val(game.InstructionResultGate{
							Key:      "disjunctive-cost-a",
							Accepted: game.TriFalse,
						}),
						Optional:      true,
						PublishResult: game.ResultKey("disjunctive-cost-b"),
					},
					{
						Primitive: game.Draw{
							Amount: game.Fixed(2),
							Player: game.ControllerReference(),
						},
						ResultGate: opt.Val(game.InstructionResultGate{
							Key:       "disjunctive-cost-a",
							Succeeded: game.TriTrue,
						}),
					},
					{
						Primitive: game.Draw{
							Amount: game.Fixed(2),
							Player: game.ControllerReference(),
						},
						ResultGate: opt.Val(game.InstructionResultGate{
							Key:       "disjunctive-cost-b",
							Succeeded: game.TriTrue,
						}),
					},
				},
			}.Ability()),
			OracleText: `
			You may discard a card or sacrifice a land. If you do, draw two cards.
			Plot {1}{R} (You may pay {1}{R} and exile this card from your hand. Cast it as a sorcery on a later turn without paying its mana cost. Plot only as a sorcery.)
		`,
		},
	}
}
