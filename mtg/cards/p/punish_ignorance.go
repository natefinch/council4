package p

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// PunishIgnorance is the card definition for Punish Ignorance.
//
// Type: Instant
// Cost: {W}{U}{U}{B}
//
// Oracle text:
//
//	Counter target spell. Its controller loses 3 life and you gain 3 life.
var PunishIgnorance = newPunishIgnorance

func newPunishIgnorance() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White, color.Blue, color.Black),
		CardFace: game.CardFace{
			Name: "Punish Ignorance",
			ManaCost: opt.Val(cost.Mana{
				cost.W,
				cost.U,
				cost.U,
				cost.B,
			}),
			Colors: []color.Color{color.Black, color.Blue, color.White},
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
					{
						Primitive: game.GainLife{
							Amount: game.Fixed(3),
							Player: game.ControllerReference(),
						},
					},
				},
			}.Ability()),
			OracleText: `
			Counter target spell. Its controller loses 3 life and you gain 3 life.
		`,
		},
	}
}
