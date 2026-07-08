package v

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// ViewFromAbove is the card definition for View from Above.
//
// Type: Instant
// Cost: {1}{U}
//
// Oracle text:
//
//	Target creature gains flying until end of turn. If you control a white permanent, return View from Above to its owner's hand.
var ViewFromAbove = newViewFromAbove

func newViewFromAbove() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue),
		CardFace: game.CardFace{
			Name: "View from Above",
			ManaCost: opt.Val(cost.Mana{
				cost.O(1),
				cost.U,
			}),
			Colors: []color.Color{color.Blue},
			Types:  []types.Card{types.Instant},
			SpellAbility: opt.Val(game.Mode{
				Targets: []game.TargetSpec{
					game.TargetSpec{
						MinTargets: 1,
						MaxTargets: 1,
						Constraint: "target creature",
						Allow:      game.TargetAllowPermanent,
						Selection:  opt.Val(game.Selection{RequiredTypesAny: []types.Card{types.Creature}}),
					},
				},
				Sequence: []game.Instruction{
					{
						Primitive: game.ApplyContinuous{
							Object: opt.Val(game.TargetPermanentReference(0)),
							ContinuousEffects: []game.ContinuousEffect{
								game.ContinuousEffect{
									Layer: game.LayerAbility,
									AddKeywords: []game.Keyword{
										game.Flying,
									},
								},
							},
							Duration: game.DurationUntilEndOfTurn,
						},
					},
					{
						Primitive: game.Bounce{
							Object: game.SourcePermanentReference(),
						},
						Condition: opt.Val(game.EffectCondition{
							Condition: opt.Val(game.Condition{
								ControlsMatching: opt.Val(game.SelectionCount{
									Selection: game.Selection{ColorsAny: []color.Color{color.White}},
								}),
							}),
						}),
					},
					{
						Primitive: game.Bounce{
							Object: game.TargetPermanentReference(0),
						},
						Condition: opt.Val(game.EffectCondition{
							Condition: opt.Val(game.Condition{
								ControlsMatching: opt.Val(game.SelectionCount{
									Selection: game.Selection{ColorsAny: []color.Color{color.White}},
								}),
							}),
						}),
					},
				},
			}.Ability()),
			OracleText: `
			Target creature gains flying until end of turn. If you control a white permanent, return View from Above to its owner's hand.
		`,
		},
	}
}
