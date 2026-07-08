package i

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// InspiringBard is the card definition for Inspiring Bard.
//
// Type: Creature — Elf Bard
// Cost: {3}{G}
//
// Oracle text:
//
//	When this creature enters, choose one —
//	• Bardic Inspiration — Target creature gets +2/+2 until end of turn.
//	• Song of Rest — You gain 3 life.
var InspiringBard = newInspiringBard

func newInspiringBard() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Green),
		CardFace: game.CardFace{
			Name: "Inspiring Bard",
			ManaCost: opt.Val(cost.Mana{
				cost.O(3),
				cost.G,
			}),
			Colors:    []color.Color{color.Green},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Elf, types.Bard},
			Power:     opt.Val(game.PT{Value: 3}),
			Toughness: opt.Val(game.PT{Value: 3}),
			TriggeredAbilities: []game.TriggeredAbility{
				game.TriggeredAbility{
					Trigger: game.TriggerCondition{
						Type: game.TriggerWhen,
						Pattern: game.TriggerPattern{
							Event:  game.EventPermanentEnteredBattlefield,
							Source: game.TriggerSourceSelf,
						},
					},
					Content: game.AbilityContent{
						Modes: []game.Mode{
							game.Mode{
								Text: "Bardic Inspiration — Target creature gets +2/+2 until end of turn.",
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
											PowerDelta:     game.Fixed(2),
											ToughnessDelta: game.Fixed(2),
											Duration:       game.DurationUntilEndOfTurn,
										},
									},
								},
							},
							game.Mode{
								Text: "Song of Rest — You gain 3 life.",
								Sequence: []game.Instruction{
									{
										Primitive: game.GainLife{
											Amount: game.Fixed(3),
											Player: game.ControllerReference(),
										},
									},
								},
							},
						},
						MinModes: 1,
						MaxModes: 1,
					},
				},
			},
			OracleText: `
			When this creature enters, choose one —
			• Bardic Inspiration — Target creature gets +2/+2 until end of turn.
			• Song of Rest — You gain 3 life.
		`,
		},
	}
}
