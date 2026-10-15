package t

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// ThreeStepsAhead is the card definition for Three Steps Ahead.
//
// Type: Instant
// Cost: {U}
//
// Oracle text:
//
//	Spree (Choose one or more additional costs.)
//	+ {1}{U} — Counter target spell.
//	+ {3} — Create a token that's a copy of target artifact or creature you control.
//	+ {2} — Draw two cards, then discard a card.
var ThreeStepsAhead = newThreeStepsAhead

func newThreeStepsAhead() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue),
		CardFace: game.CardFace{
			Name: "Three Steps Ahead",
			ManaCost: opt.Val(cost.Mana{
				cost.U,
			}),
			Colors: []color.Color{color.Blue},
			Types:  []types.Card{types.Instant},
			SpellAbility: opt.Val(game.AbilityContent{
				Modes: []game.Mode{
					game.Mode{
						Text: "{1}{U} — Counter target spell.",
						Cost: opt.Val(cost.Mana{cost.O(1), cost.U}),
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
					},
					game.Mode{
						Text: "{3} — Create a token that's a copy of target artifact or creature you control.",
						Cost: opt.Val(cost.Mana{cost.O(3)}),
						Targets: []game.TargetSpec{
							game.TargetSpec{
								MinTargets: 1,
								MaxTargets: 1,
								Constraint: "target artifact or creature you control",
								Allow:      game.TargetAllowPermanent,
								Selection:  opt.Val(game.Selection{RequiredTypesAny: []types.Card{types.Artifact, types.Creature}, Controller: game.ControllerYou}),
							},
						},
						Sequence: []game.Instruction{
							{
								Primitive: game.CreateToken{
									Amount: game.Fixed(1),
									Source: game.TokenCopyOf(game.TokenCopySpec{
										Source: game.TokenCopySourceObject,
										Object: game.TargetPermanentReference(0),
									}),
								},
							},
						},
					},
					game.Mode{
						Text: "{2} — Draw two cards, then discard a card.",
						Cost: opt.Val(cost.Mana{cost.O(2)}),
						Sequence: []game.Instruction{
							{
								Primitive: game.Draw{
									Amount: game.Fixed(2),
									Player: game.ControllerReference(),
								},
							},
							{
								Primitive: game.Discard{
									Amount: game.Fixed(1),
									Player: game.ControllerReference(),
								},
							},
						},
					},
				},
				MinModes: 1,
				MaxModes: 3,
			}),
			OracleText: `
			Spree (Choose one or more additional costs.)
			+ {1}{U} — Counter target spell.
			+ {3} — Create a token that's a copy of target artifact or creature you control.
			+ {2} — Draw two cards, then discard a card.
		`,
		},
	}
}
