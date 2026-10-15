package k

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// KissOfTheAmesha is the card definition for Kiss of the Amesha.
//
// Type: Sorcery
// Cost: {4}{W}{U}
//
// Oracle text:
//
//	Target player gains 7 life and draws two cards.
var KissOfTheAmesha = newKissOfTheAmesha

func newKissOfTheAmesha() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White, color.Blue),
		CardFace: game.CardFace{
			Name: "Kiss of the Amesha",
			ManaCost: opt.Val(cost.Mana{
				cost.O(4),
				cost.W,
				cost.U,
			}),
			Colors: []color.Color{color.Blue, color.White},
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
						Primitive: game.GainLife{
							Amount: game.Fixed(7),
							Player: game.TargetPlayerReference(0),
						},
					},
					{
						Primitive: game.Draw{
							Amount: game.Fixed(2),
							Player: game.TargetPlayerReference(0),
						},
					},
				},
			}.Ability()),
			OracleText: `
			Target player gains 7 life and draws two cards.
		`,
		},
	}
}
