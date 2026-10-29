package s

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// Shunt is the card definition for Shunt.
//
// Type: Instant
// Cost: {1}{R}{R}
//
// Oracle text:
//
//	Change the target of target spell with a single target.
var Shunt = newShunt

func newShunt() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Red),
		CardFace: game.CardFace{
			Name: "Shunt",
			ManaCost: opt.Val(cost.Mana{
				cost.O(1),
				cost.R,
				cost.R,
			}),
			Colors: []color.Color{color.Red},
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
							SingleTarget:     true,
						},
					},
				},
				Sequence: []game.Instruction{
					{
						Primitive: game.ChooseNewTargets{
							Object: game.TargetStackObjectReference(0),
						},
					},
				},
			}.Ability()),
			OracleText: `
			Change the target of target spell with a single target.
		`,
		},
	}
}
