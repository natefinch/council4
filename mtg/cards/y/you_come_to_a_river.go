package y

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// YouComeToARiver is the card definition for You Come to a River.
//
// Type: Instant
// Cost: {1}{U}
//
// Oracle text:
//
//	Choose one —
//	• Fight the Current — Return target nonland permanent to its owner's hand.
//	• Find a Crossing — Target creature gets +1/+0 until end of turn and can't be blocked this turn.
var YouComeToARiver = newYouComeToARiver()

func newYouComeToARiver() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue),
		CardFace: game.CardFace{
			Name: "You Come to a River",
			ManaCost: opt.Val(cost.Mana{
				cost.O(1),
				cost.U,
			}),
			Colors: []color.Color{color.Blue},
			Types:  []types.Card{types.Instant},
			SpellAbility: opt.Val(game.AbilityContent{
				Modes: []game.Mode{
					game.Mode{
						Text: "Fight the Current — Return target nonland permanent to its owner's hand.",
						Targets: []game.TargetSpec{
							game.TargetSpec{
								MinTargets: 1,
								MaxTargets: 1,
								Constraint: "target nonland permanent",
								Allow:      game.TargetAllowPermanent,
								Selection:  opt.Val(game.Selection{ExcludedTypes: []types.Card{types.Land}}),
							},
						},
						Sequence: []game.Instruction{
							{
								Primitive: game.Bounce{
									Object: game.TargetPermanentReference(0),
								},
							},
						},
					},
					game.Mode{
						Text: "Find a Crossing — Target creature gets +1/+0 until end of turn and can't be blocked this turn.",
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
								Primitive: game.ModifyPT{
									Object:         game.TargetPermanentReference(0),
									PowerDelta:     game.Fixed(1),
									ToughnessDelta: game.Fixed(0),
									Duration:       game.DurationUntilEndOfTurn,
								},
							},
							{
								Primitive: game.ApplyRule{
									Object: opt.Val(game.TargetPermanentReference(0)),
									RuleEffects: []game.RuleEffect{
										game.RuleEffect{
											Kind: game.RuleEffectCantBeBlocked,
										},
									},
									Duration: game.DurationThisTurn,
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
			• Fight the Current — Return target nonland permanent to its owner's hand.
			• Find a Crossing — Target creature gets +1/+0 until end of turn and can't be blocked this turn.
		`,
		},
	}
}
