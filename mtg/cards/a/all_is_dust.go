package a

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// AllIsDust is the card definition for All Is Dust.
//
// Type: Kindred Sorcery — Eldrazi
// Cost: {7}
//
// Oracle text:
//
//	Each player sacrifices all permanents they control that are one or more colors.
var AllIsDust = newAllIsDust()

func newAllIsDust() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name: "All Is Dust",
			ManaCost: opt.Val(cost.Mana{
				cost.O(7),
			}),
			Types:    []types.Card{types.Kindred, types.Sorcery},
			Subtypes: []types.Sub{types.Eldrazi},
			SpellAbility: opt.Val(game.Mode{
				Sequence: []game.Instruction{
					{
						Primitive: game.SacrificePermanents{
							All:         true,
							PlayerGroup: game.AllPlayersReference(),
							Selection:   game.Selection{Colored: true},
						},
					},
				},
			}.Ability()),
			OracleText: `
			Each player sacrifices all permanents they control that are one or more colors.
		`,
		},
	}
}
