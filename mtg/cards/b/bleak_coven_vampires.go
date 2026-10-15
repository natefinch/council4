package b

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// BleakCovenVampires is the card definition for Bleak Coven Vampires.
//
// Type: Creature — Vampire Warrior
// Cost: {3}{B}{B}
//
// Oracle text:
//
//	Metalcraft — When this creature enters, if you control three or more artifacts, target player loses 4 life and you gain 4 life.
var BleakCovenVampires = newBleakCovenVampires

func newBleakCovenVampires() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Black),
		CardFace: game.CardFace{
			Name: "Bleak Coven Vampires",
			ManaCost: opt.Val(cost.Mana{
				cost.O(3),
				cost.B,
				cost.B,
			}),
			Colors:    []color.Color{color.Black},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Vampire, types.Warrior},
			Power:     opt.Val(game.PT{Value: 4}),
			Toughness: opt.Val(game.PT{Value: 3}),
			TriggeredAbilities: []game.TriggeredAbility{
				game.TriggeredAbility{
					Text: "Metalcraft — When this creature enters, if you control three or more artifacts, target player loses 4 life and you gain 4 life.",
					Trigger: game.TriggerCondition{
						Type: game.TriggerWhen,
						Pattern: game.TriggerPattern{
							Event:           game.EventPermanentEnteredBattlefield,
							Source:          game.TriggerSourceSelf,
							StackObjectKind: game.StackSpell,
							CounterKind:     counter.PlusOnePlusOne,
						},
						InterveningIf: "if you control three or more artifacts",
						InterveningCondition: opt.Val(game.Condition{
							Text: "if you control three or more artifacts",
							ControlsMatching: opt.Val(game.SelectionCount{
								Selection: game.Selection{RequiredTypes: []types.Card{types.Artifact}},
								MinCount:  3,
							}),
							SourceCounterKind: counter.PlusOnePlusOne,
						}),
					},
					Content: game.AbilityContent{
						Modes: []game.Mode{
							game.Mode{
								Targets: []game.TargetSpec{
									game.TargetSpec{
										MinTargets: 1,
										MaxTargets: 1,
										Constraint: "target player",
										Allow:      game.TargetAllowPlayer,
									},
								},
								Sequence: []game.Instruction{
									game.Instruction{
										Primitive: game.LoseLife{
											Amount: game.Fixed(4),
											Player: game.TargetPlayerReference(0),
										},
									},
									game.Instruction{
										Primitive: game.GainLife{
											Amount: game.Fixed(4),
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
			Metalcraft — When this creature enters, if you control three or more artifacts, target player loses 4 life and you gain 4 life.
		`,
		},
	}
}
