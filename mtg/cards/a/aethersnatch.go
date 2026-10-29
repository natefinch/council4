package a

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// Aethersnatch is the card definition for Aethersnatch.
//
// Type: Instant
// Cost: {4}{U}{U}
//
// Oracle text:
//
//	Gain control of target spell. You may choose new targets for it. (If that spell becomes a permanent, it enters under your control.)
var Aethersnatch = newAethersnatch

func newAethersnatch() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue),
		CardFace: game.CardFace{
			Name: "Aethersnatch",
			ManaCost: opt.Val(cost.Mana{
				cost.O(4),
				cost.U,
				cost.U,
			}),
			Colors: []color.Color{color.Blue},
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
						Primitive: game.ChangeStackObjectController{
							Object:     game.TargetStackObjectReference(0),
							Controller: game.ControllerReference(),
						},
					},
					{
						Primitive: game.ChooseNewTargets{
							Object: game.TargetStackObjectReference(0),
							Rule:   game.RetargetChooseNewTargets,
						},
						Optional: true,
					},
				},
			}.Ability()),
			OracleText: `
			Gain control of target spell. You may choose new targets for it. (If that spell becomes a permanent, it enters under your control.)
		`,
		},
	}
}
