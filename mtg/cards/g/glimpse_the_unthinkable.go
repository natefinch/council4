package g

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// GlimpseTheUnthinkable is the card definition for Glimpse the Unthinkable.
//
// Type: Sorcery
// Cost: {U}{B}
//
// Oracle text:
//
//	Target player mills ten cards.
var GlimpseTheUnthinkable = newGlimpseTheUnthinkable

func newGlimpseTheUnthinkable() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue, color.Black),
		CardFace: game.CardFace{
			Name: "Glimpse the Unthinkable",
			ManaCost: opt.Val(cost.Mana{
				cost.U,
				cost.B,
			}),
			Colors: []color.Color{color.Black, color.Blue},
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
						Primitive: game.Mill{
							Amount: game.Fixed(10),
							Player: game.TargetPlayerReference(0),
						},
					},
				},
			}.Ability()),
			OracleText: `
			Target player mills ten cards.
		`,
		},
	}
}
