package t

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// Thwart is the card definition for Thwart.
//
// Type: Instant
// Cost: {2}{U}{U}
//
// Oracle text:
//
//	You may return three Islands you control to their owner's hand rather than pay this spell's mana cost.
//	Counter target spell.
var Thwart = newThwart

func newThwart() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue),
		CardFace: game.CardFace{
			Name: "Thwart",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
				cost.U,
				cost.U,
			}),
			Colors: []color.Color{color.Blue},
			Types:  []types.Card{types.Instant},
			AlternativeCosts: []cost.Alternative{
				cost.Alternative{
					Label: "Return three Islands you control to their owner's hand",
					AdditionalCosts: []cost.Additional{
						{
							Kind:        cost.AdditionalReturnToHand,
							Text:        "return three Islands you control to their owner's hand",
							Amount:      3,
							SubtypesAny: cost.SubtypeSet{types.Island},
						},
					},
				},
			},
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
				},
			}.Ability()),
			OracleText: `
			You may return three Islands you control to their owner's hand rather than pay this spell's mana cost.
			Counter target spell.
		`,
		},
	}
}
