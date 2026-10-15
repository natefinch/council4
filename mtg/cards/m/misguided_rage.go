package m

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// MisguidedRage is the card definition for Misguided Rage.
//
// Type: Sorcery
// Cost: {2}{R}
//
// Oracle text:
//
//	Target player sacrifices a permanent of their choice.
var MisguidedRage = newMisguidedRage

func newMisguidedRage() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Red),
		CardFace: game.CardFace{
			Name: "Misguided Rage",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
				cost.R,
			}),
			Colors: []color.Color{color.Red},
			Types:  []types.Card{types.Sorcery},
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
							Player: game.TargetPlayerReference(0),
							Amount: game.Fixed(1),
						},
					},
				},
			}.Ability()),
			OracleText: `
			Target player sacrifices a permanent of their choice.
		`,
		},
	}
}
