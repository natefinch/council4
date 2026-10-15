package u

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// Undermine is the card definition for Undermine.
//
// Type: Instant
// Cost: {U}{U}{B}
//
// Oracle text:
//
//	Counter target spell. Its controller loses 3 life.
var Undermine = newUndermine

func newUndermine() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue, color.Black),
		CardFace: game.CardFace{
			Name: "Undermine",
			ManaCost: opt.Val(cost.Mana{
				cost.U,
				cost.U,
				cost.B,
			}),
			Colors: []color.Color{color.Black, color.Blue},
			Types:  []types.Card{types.Instant},
			SpellAbility: opt.Val(game.Mode{
				Targets: []game.TargetSpec{
					game.TargetSpec{
						MinTargets: 1,
						MaxTargets: 1,
						Constraint: "target spell",
						Allow:      game.TargetAllowStackObject,
						Predicate: game.TargetPredicate{
							StackObjectKinds: []game.StackObjectKind{game.StackSpell},
						},
					},
				},
				Sequence: []game.Instruction{
					{
						Primitive: game.CounterObject{
							Object: game.TargetStackObjectReference(0),
						},
					},
					{
						Primitive: game.LoseLife{
							Amount: game.Fixed(3),
							Player: game.ObjectControllerReference(game.TargetStackObjectReference(0)),
						},
					},
				},
			}.Ability()),
			OracleText: `
			Counter target spell. Its controller loses 3 life.
		`,
		},
	}
}
