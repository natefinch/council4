package o

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// OperaLoveSong is the card definition for Opera Love Song.
//
// Type: Instant
// Cost: {1}{R}
//
// Oracle text:
//
//	Choose one —
//	• Exile the top two cards of your library. You may play those cards until your next end step.
//	• One or two target creatures each get +2/+0 until end of turn.
var OperaLoveSong = newOperaLoveSong()

func newOperaLoveSong() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Red),
		CardFace: game.CardFace{
			Name: "Opera Love Song",
			ManaCost: opt.Val(cost.Mana{
				cost.O(1),
				cost.R,
			}),
			Colors: []color.Color{color.Red},
			Types:  []types.Card{types.Instant},
			SpellAbility: opt.Val(game.AbilityContent{
				Modes: []game.Mode{
					game.Mode{
						Text: "Exile the top two cards of your library. You may play those cards until your next end step.",
						Sequence: []game.Instruction{
							{
								Primitive: game.ImpulseExile{
									Player:   game.ControllerReference(),
									Amount:   game.Fixed(2),
									Duration: game.DurationUntilYourNextEndStep,
								},
							},
						},
					},
					game.Mode{
						Text: "One or two target creatures each get +2/+0 until end of turn.",
						Targets: []game.TargetSpec{
							game.TargetSpec{
								MinTargets: 1,
								MaxTargets: 2,
								Constraint: "one or two target creatures",
								Allow:      game.TargetAllowPermanent,
								Selection:  opt.Val(game.Selection{RequiredTypesAny: []types.Card{types.Creature}}),
							},
						},
						Sequence: []game.Instruction{
							{
								Primitive: game.ModifyPT{
									Object:         game.TargetPermanentReference(0),
									PowerDelta:     game.Fixed(2),
									ToughnessDelta: game.Fixed(0),
									Duration:       game.DurationUntilEndOfTurn,
								},
							},
							{
								Primitive: game.ModifyPT{
									Object:         game.TargetPermanentReference(1),
									PowerDelta:     game.Fixed(2),
									ToughnessDelta: game.Fixed(0),
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
			• Exile the top two cards of your library. You may play those cards until your next end step.
			• One or two target creatures each get +2/+0 until end of turn.
		`,
		},
	}
}
