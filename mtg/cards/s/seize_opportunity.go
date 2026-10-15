package s

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// SeizeOpportunity is the card definition for Seize Opportunity.
//
// Type: Instant
// Cost: {2}{R}
//
// Oracle text:
//
//	Choose one —
//	• Exile the top two cards of your library. Until the end of your next turn, you may play those cards.
//	• Up to two target creatures each get +2/+1 until end of turn.
var SeizeOpportunity = newSeizeOpportunity

func newSeizeOpportunity() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Red),
		CardFace: game.CardFace{
			Name: "Seize Opportunity",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
				cost.R,
			}),
			Colors: []color.Color{color.Red},
			Types:  []types.Card{types.Instant},
			SpellAbility: opt.Val(game.AbilityContent{
				Modes: []game.Mode{
					game.Mode{
						Text: "Exile the top two cards of your library. Until the end of your next turn, you may play those cards.",
						Sequence: []game.Instruction{
							{
								Primitive: game.ImpulseExile{
									Player:   game.ControllerReference(),
									Amount:   game.Fixed(2),
									Duration: game.DurationUntilEndOfYourNextTurn,
								},
							},
						},
					},
					game.Mode{
						Text: "Up to two target creatures each get +2/+1 until end of turn.",
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
								Primitive: game.ModifyPT{
									Object:         game.TargetPermanentReference(0),
									PowerDelta:     game.Fixed(2),
									ToughnessDelta: game.Fixed(1),
									Duration:       game.DurationUntilEndOfTurn,
								},
							},
							{
								Primitive: game.ModifyPT{
									Object:         game.TargetPermanentReference(1),
									PowerDelta:     game.Fixed(2),
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
			• Exile the top two cards of your library. Until the end of your next turn, you may play those cards.
			• Up to two target creatures each get +2/+1 until end of turn.
		`,
		},
	}
}
