package p

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// Pyroblast is the card definition for Pyroblast.
//
// Type: Instant
// Cost: {R}
//
// Oracle text:
//
//	Choose one —
//	• Counter target spell if it's blue.
//	• Destroy target permanent if it's blue.
var Pyroblast = newPyroblast

func newPyroblast() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Red),
		CardFace: game.CardFace{
			Name: "Pyroblast",
			ManaCost: opt.Val(cost.Mana{
				cost.R,
			}),
			Colors: []color.Color{color.Red},
			Types:  []types.Card{types.Instant},
			SpellAbility: opt.Val(game.AbilityContent{
				Modes: []game.Mode{
					game.Mode{
						Text: "Counter target spell if it's blue.",
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
								Condition: opt.Val(game.EffectCondition{
									Object: game.TargetStackObjectReference(0),
									Condition: opt.Val(game.Condition{
										Object:        opt.Val(game.TargetStackObjectReference(0)),
										ObjectMatches: opt.Val(game.Selection{ColorsAny: []color.Color{color.Blue}}),
									}),
								}),
							},
						},
					},
					game.Mode{
						Text: "Destroy target permanent if it's blue.",
						Targets: []game.TargetSpec{
							game.TargetSpec{
								MinTargets: 1,
								MaxTargets: 1,
								Constraint: "target permanent",
								Allow:      game.TargetAllowPermanent,
							},
						},
						Sequence: []game.Instruction{
							{
								Primitive: game.Destroy{
									Object: game.TargetPermanentReference(0),
								},
								Condition: opt.Val(game.EffectCondition{
									Object: game.TargetPermanentReference(0),
									Condition: opt.Val(game.Condition{
										Object:        opt.Val(game.TargetPermanentReference(0)),
										ObjectMatches: opt.Val(game.Selection{ColorsAny: []color.Color{color.Blue}}),
									}),
								}),
							},
						},
					},
				},
				MinModes: 1,
				MaxModes: 1,
			}),
			OracleText: `
			Choose one —
			• Counter target spell if it's blue.
			• Destroy target permanent if it's blue.
		`,
		},
	}
}
