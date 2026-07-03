package p

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// ProudPackRhino is the card definition for Proud Pack-Rhino.
//
// Type: Creature — Rhino
// Cost: {2}{W}
//
// Oracle text:
//
//	When this creature enters, choose one —
//	• Put a shield counter on target permanent. (If it would be dealt damage or destroyed, remove a shield counter from it instead.)
//	• Proliferate. (Choose any number of permanents and/or players, then give each another counter of each kind already there.)
var ProudPackRhino = newProudPackRhino()

func newProudPackRhino() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White),
		CardFace: game.CardFace{
			Name: "Proud Pack-Rhino",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
				cost.W,
			}),
			Colors:    []color.Color{color.White},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Rhino},
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
								Text: "Put a shield counter on target permanent. (If it would be dealt damage or destroyed, remove a shield counter from it instead.)",
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
										Primitive: game.AddCounter{
											Amount:      game.Fixed(1),
											Object:      game.TargetPermanentReference(0),
											CounterKind: counter.Shield,
										},
									},
								},
							},
							game.Mode{
								Text: "Proliferate. (Choose any number of permanents and/or players, then give each another counter of each kind already there.)",
								Sequence: []game.Instruction{
									{
										Primitive: game.Proliferate{
											Amount: game.Fixed(1),
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
			• Put a shield counter on target permanent. (If it would be dealt damage or destroyed, remove a shield counter from it instead.)
			• Proliferate. (Choose any number of permanents and/or players, then give each another counter of each kind already there.)
		`,
		},
	}
}
