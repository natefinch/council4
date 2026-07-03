package g

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// GethSVerdict is the card definition for Geth's Verdict.
//
// Type: Instant
// Cost: {B}{B}
//
// Oracle text:
//
//	Target player sacrifices a creature of their choice and loses 1 life.
var GethSVerdict = newGethSVerdict()

func newGethSVerdict() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Black),
		CardFace: game.CardFace{
			Name: "Geth's Verdict",
			ManaCost: opt.Val(cost.Mana{
				cost.B,
				cost.B,
			}),
			Colors: []color.Color{color.Black},
			Types:  []types.Card{types.Instant},
			SpellAbility: opt.Val(game.Mode{
				Targets: []game.TargetSpec{
					game.TargetSpec{
						MinTargets: 1,
						MaxTargets: 1,
						Constraint: "Target player",
						Allow:      game.TargetAllowPlayer,
					},
				},
				Sequence: []game.Instruction{
					{
						Primitive: game.SacrificePermanents{
							Amount:    game.Fixed(1),
							Player:    game.TargetPlayerReference(0),
							Selection: game.Selection{RequiredTypes: []types.Card{types.Creature}},
						},
					},
					{
						Primitive: game.LoseLife{
							Amount: game.Fixed(1),
							Player: game.TargetPlayerReference(0),
						},
					},
				},
			}.Ability()),
			OracleText: `
			Target player sacrifices a creature of their choice and loses 1 life.
		`,
		},
	}
}
