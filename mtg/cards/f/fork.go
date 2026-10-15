package f

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// Fork is the card definition for Fork.
//
// Type: Instant
// Cost: {R}{R}
//
// Oracle text:
//
//	Copy target instant or sorcery spell, except that the copy is red. You may choose new targets for the copy.
var Fork = newFork

func newFork() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Red),
		CardFace: game.CardFace{
			Name: "Fork",
			ManaCost: opt.Val(cost.Mana{
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
						Constraint: "target instant or sorcery spell",
						Allow:      game.TargetAllowStackObject,
						Predicate: game.TargetPredicate{
							SpellCardTypesAny: []types.Card{types.Instant, types.Sorcery},
							StackObjectKinds:  []game.StackObjectKind{game.StackSpell},
						},
					},
				},
				Sequence: []game.Instruction{
					{
						Primitive: game.CopyStackObject{
							Object:              game.TargetStackObjectReference(0),
							Count:               1,
							MayChooseNewTargets: true,
							SetColors:           []color.Color{color.Red},
						},
					},
				},
			}.Ability()),
			OracleText: `
			Copy target instant or sorcery spell, except that the copy is red. You may choose new targets for the copy.
		`,
		},
	}
}
