package c

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// CrackTheEarth is the card definition for Crack the Earth.
//
// Type: Sorcery — Arcane
// Cost: {R}
//
// Oracle text:
//
//	Each player sacrifices a permanent of their choice.
var CrackTheEarth = newCrackTheEarth

func newCrackTheEarth() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Red),
		CardFace: game.CardFace{
			Name: "Crack the Earth",
			ManaCost: opt.Val(cost.Mana{
				cost.R,
			}),
			Colors:   []color.Color{color.Red},
			Types:    []types.Card{types.Sorcery},
			Subtypes: []types.Sub{types.Arcane},
			SpellAbility: opt.Val(game.Mode{
				Sequence: []game.Instruction{
					{
						Primitive: game.SacrificePermanents{
							PlayerGroup: game.AllPlayersReference(),
							Amount:      game.Fixed(1),
						},
					},
				},
			}.Ability()),
			OracleText: `
			Each player sacrifices a permanent of their choice.
		`,
		},
	}
}
