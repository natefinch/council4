package b

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// BlitzLeech is the card definition for Blitz Leech.
//
// Type: Creature — Leech
// Cost: {5}{B}
//
// Oracle text:
//
//	Flash
//	When this creature enters, target creature an opponent controls gets -2/-2 until end of turn. Remove all counters from that creature.
var BlitzLeech = newBlitzLeech

func newBlitzLeech() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Black),
		CardFace: game.CardFace{
			Name: "Blitz Leech",
			ManaCost: opt.Val(cost.Mana{
				cost.O(5),
				cost.B,
			}),
			Colors:    []color.Color{color.Black},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Leech},
			Power:     opt.Val(game.PT{Value: 5}),
			Toughness: opt.Val(game.PT{Value: 2}),
			StaticAbilities: []game.StaticAbility{
				game.FlashStaticBody,
			},
			TriggeredAbilities: []game.TriggeredAbility{
				game.TriggeredAbility{
					Trigger: game.TriggerCondition{
						Type: game.TriggerWhen,
						Pattern: game.TriggerPattern{
							Event:  game.EventPermanentEnteredBattlefield,
							Source: game.TriggerSourceSelf,
						},
					},
					Content: game.Mode{
						Targets: []game.TargetSpec{
							game.TargetSpec{
								MinTargets: 1,
								MaxTargets: 1,
								Constraint: "target creature an opponent controls",
								Allow:      game.TargetAllowPermanent,
								Selection:  opt.Val(game.Selection{RequiredTypesAny: []types.Card{types.Creature}, Controller: game.ControllerOpponent}),
							},
						},
						Sequence: []game.Instruction{
							{
								Primitive: game.ModifyPT{
									Object:         game.TargetPermanentReference(0),
									PowerDelta:     game.Fixed(-2),
									ToughnessDelta: game.Fixed(-2),
									Duration:       game.DurationUntilEndOfTurn,
								},
							},
							{
								Primitive: game.RemoveCounter{
									Object:   game.TargetPermanentReference(0),
									AllKinds: true,
								},
							},
						},
					}.Ability(),
				},
			},
			OracleText: `
			Flash
			When this creature enters, target creature an opponent controls gets -2/-2 until end of turn. Remove all counters from that creature.
		`,
		},
	}
}
