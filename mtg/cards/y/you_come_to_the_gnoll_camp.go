package y

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// YouComeToTheGnollCamp is the card definition for You Come to the Gnoll Camp.
//
// Type: Instant
// Cost: {1}{R}
//
// Oracle text:
//
//	Choose one —
//	• Intimidate Them — Up to two target creatures can't block this turn.
//	• Fend Them Off — Target creature gets +3/+1 until end of turn.
var YouComeToTheGnollCamp = newYouComeToTheGnollCamp

func newYouComeToTheGnollCamp() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Red),
		CardFace: game.CardFace{
			Name: "You Come to the Gnoll Camp",
			ManaCost: opt.Val(cost.Mana{
				cost.O(1),
				cost.R,
			}),
			Colors: []color.Color{color.Red},
			Types:  []types.Card{types.Instant},
			SpellAbility: opt.Val(game.AbilityContent{
				Modes: []game.Mode{
					game.Mode{
						Text: "Intimidate Them — Up to two target creatures can't block this turn.",
						Targets: []game.TargetSpec{
							game.TargetSpec{
								MinTargets: 0,
								MaxTargets: 2,
								Constraint: "up to two target creatures",
								Allow:      game.TargetAllowPermanent,
								Selection:  opt.Val(game.Selection{RequiredTypesAny: []types.Card{types.Creature}}),
							},
						},
						Sequence: []game.Instruction{
							{
								Primitive: game.ApplyRule{
									Object: opt.Val(game.TargetPermanentReference(0)),
									RuleEffects: []game.RuleEffect{
										game.RuleEffect{
											Kind:     game.RuleEffectCantBlock,
											Duration: game.DurationPermanent,
										},
									},
									Duration: game.DurationThisTurn,
								},
							},
							{
								Primitive: game.ApplyRule{
									Object: opt.Val(game.TargetPermanentReference(1)),
									RuleEffects: []game.RuleEffect{
										game.RuleEffect{
											Kind:     game.RuleEffectCantBlock,
											Duration: game.DurationPermanent,
										},
									},
									Duration: game.DurationThisTurn,
								},
							},
						},
					},
					game.Mode{
						Text: "Fend Them Off — Target creature gets +3/+1 until end of turn.",
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
									PowerDelta:     game.Fixed(3),
									ToughnessDelta: game.Fixed(1),
									Duration:       game.DurationUntilEndOfTurn,
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
			• Intimidate Them — Up to two target creatures can't block this turn.
			• Fend Them Off — Target creature gets +3/+1 until end of turn.
		`,
		},
	}
}
