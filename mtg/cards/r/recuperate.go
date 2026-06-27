package r

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// Recuperate is the card definition for Recuperate.
//
// Type: Instant
// Cost: {3}{W}
//
// Oracle text:
//
//	Choose one —
//	• You gain 6 life.
//	• Prevent the next 6 damage that would be dealt to target creature this turn.
var Recuperate = newRecuperate()

func newRecuperate() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White),
		CardFace: game.CardFace{
			Name: "Recuperate",
			ManaCost: opt.Val(cost.Mana{
				cost.O(3),
				cost.W,
			}),
			Colors: []color.Color{color.White},
			Types:  []types.Card{types.Instant},
			SpellAbility: opt.Val(game.AbilityContent{
				Modes: []game.Mode{
					game.Mode{
						Text: "You gain 6 life.",
						Sequence: []game.Instruction{
							{
								Primitive: game.GainLife{
									Amount: game.Fixed(6),
									Player: game.ControllerReference(),
								},
							},
						},
					},
					game.Mode{
						Text: "Prevent the next 6 damage that would be dealt to target creature this turn.",
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
								Primitive: game.PreventDamage{
									AnyTarget: game.AnyTargetDamageRecipient(0),
									Amount:    game.Fixed(6),
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
			• You gain 6 life.
			• Prevent the next 6 damage that would be dealt to target creature this turn.
		`,
		},
	}
}
