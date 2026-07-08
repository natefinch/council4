package g

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// GlacialGrasp is the card definition for Glacial Grasp.
//
// Type: Instant
// Cost: {2}{U}
//
// Oracle text:
//
//	Tap target creature. Its controller mills two cards. That creature doesn't untap during its controller's next untap step. (They put the top two cards of their library into their graveyard.)
//	Draw a card.
var GlacialGrasp = newGlacialGrasp

func newGlacialGrasp() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue),
		CardFace: game.CardFace{
			Name: "Glacial Grasp",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
				cost.U,
			}),
			Colors: []color.Color{color.Blue},
			Types:  []types.Card{types.Instant},
			SpellAbility: opt.Val(game.Mode{
				Targets: []game.TargetSpec{
					game.TargetSpec{
						MinTargets: 1,
						MaxTargets: 1,
						Constraint: "target creature",
						Allow:      game.TargetAllowPermanent,
						Selection:  opt.Val(game.Selection{RequiredTypesAny: []types.Card{types.Creature}}),
					},
				},
				Sequence: []game.Instruction{
					{
						Primitive: game.Tap{
							Object: game.TargetPermanentReference(0),
						},
					},
					{
						Primitive: game.Mill{
							Amount: game.Fixed(2),
							Player: game.ObjectControllerReference(game.TargetPermanentReference(0)),
						},
					},
					{
						Primitive: game.SkipNextUntap{
							Object: game.TargetPermanentReference(0),
						},
					},
					{
						Primitive: game.Draw{
							Amount: game.Fixed(1),
							Player: game.ControllerReference(),
						},
					},
				},
			}.Ability()),
			OracleText: `
			Tap target creature. Its controller mills two cards. That creature doesn't untap during its controller's next untap step. (They put the top two cards of their library into their graveyard.)
			Draw a card.
		`,
		},
	}
}
