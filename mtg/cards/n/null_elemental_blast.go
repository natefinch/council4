package n

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// NullElementalBlast is the card definition for Null Elemental Blast.
//
// Type: Instant
// Cost: {C}
//
// Oracle text:
//
//	Choose one —
//	• Counter target multicolored spell.
//	• Destroy target multicolored permanent.
var NullElementalBlast = newNullElementalBlast()

func newNullElementalBlast() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name: "Null Elemental Blast",
			ManaCost: opt.Val(cost.Mana{
				cost.C,
			}),
			Types: []types.Card{types.Instant},
			SpellAbility: opt.Val(game.AbilityContent{
				Modes: []game.Mode{
					game.Mode{
						Text: "Counter target multicolored spell.",
						Targets: []game.TargetSpec{
							game.TargetSpec{
								MinTargets: 1,
								MaxTargets: 1,
								Constraint: "target multicolored spell",
								Allow:      game.TargetAllowStackObject,
								Predicate: game.TargetPredicate{
									StackObjectKinds:  []game.StackObjectKind{game.StackSpell},
									SpellMulticolored: true,
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
					},
					game.Mode{
						Text: "Destroy target multicolored permanent.",
						Targets: []game.TargetSpec{
							game.TargetSpec{
								MinTargets: 1,
								MaxTargets: 1,
								Constraint: "target multicolored permanent",
								Allow:      game.TargetAllowPermanent,
								Selection:  opt.Val(game.Selection{Multicolored: true}),
							},
						},
						Sequence: []game.Instruction{
							{
								Primitive: game.Destroy{
									Object: game.TargetPermanentReference(0),
								},
							},
						},
					},
				},
				MinModes: 1,
				MaxModes: 1,
			}),
			OracleText: `
			Choose one —
			• Counter target multicolored spell.
			• Destroy target multicolored permanent.
		`,
		},
	}
}
