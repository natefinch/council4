package s

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// SpecterSWail is the card definition for Specter's Wail.
//
// Type: Sorcery
// Cost: {1}{B}
//
// Oracle text:
//
//	Target player discards a card at random.
var SpecterSWail = newSpecterSWail()

func newSpecterSWail() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Black),
		CardFace: game.CardFace{
			Name: "Specter's Wail",
			ManaCost: opt.Val(cost.Mana{
				cost.O(1),
				cost.B,
			}),
			Colors: []color.Color{color.Black},
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
						Primitive: game.Discard{
							Amount:   game.Fixed(1),
							Player:   game.TargetPlayerReference(0),
							AtRandom: true,
						},
					},
				},
			}.Ability()),
			OracleText: `
			Target player discards a card at random.
		`,
		},
	}
}
