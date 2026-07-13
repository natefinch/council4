package n

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// NarsetSReversal is the card definition for Narset's Reversal.
//
// Type: Instant
// Cost: {U}{U}
//
// Oracle text:
//
//	Copy target instant or sorcery spell, then return it to its owner's hand. You may choose new targets for the copy.
var NarsetSReversal = newNarsetSReversal

func newNarsetSReversal() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue),
		CardFace: game.CardFace{
			Name: "Narset's Reversal",
			ManaCost: opt.Val(cost.Mana{
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
							MayChooseNewTargets: true,
						},
					},
					{
						Primitive: game.Bounce{
							Object: game.TargetObjectReference(0),
						},
					},
				},
			}.Ability()),
			OracleText: `
			Copy target instant or sorcery spell, then return it to its owner's hand. You may choose new targets for the copy.
		`,
		},
	}
}
