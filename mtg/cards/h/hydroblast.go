package h

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// Hydroblast is the card definition for Hydroblast.
//
// Type: Instant
// Cost: {U}
//
// Oracle text:
//
//	Choose one —
//	• Counter target spell if it's red.
//	• Destroy target permanent if it's red.
var Hydroblast = newHydroblast

func newHydroblast() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue),
		CardFace: game.CardFace{
			Name: "Hydroblast",
			ManaCost: opt.Val(cost.Mana{
				cost.U,
			}),
			Colors: []color.Color{color.Blue},
			Types:  []types.Card{types.Instant},
			SpellAbility: opt.Val(game.AbilityContent{
				Modes: []game.Mode{
					game.Mode{
						Text: "Counter target spell if it's red.",
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
										ObjectMatches: opt.Val(game.Selection{ColorsAny: []color.Color{color.Red}}),
									}),
								}),
							},
						},
					},
					game.Mode{
						Text: "Destroy target permanent if it's red.",
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
										ObjectMatches: opt.Val(game.Selection{ColorsAny: []color.Color{color.Red}}),
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
			• Counter target spell if it's red.
			• Destroy target permanent if it's red.
		`,
		},
	}
}
