package m

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// Mournwhelk is the card definition for Mournwhelk.
//
// Type: Creature — Elemental
// Cost: {6}{B}
//
// Oracle text:
//
//	When this creature enters, target player discards two cards.
//	Evoke {3}{B} (You may cast this spell for its evoke cost. If you do, it's sacrificed when it enters.)
var Mournwhelk = newMournwhelk

func newMournwhelk() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Black),
		CardFace: game.CardFace{
			Name: "Mournwhelk",
			ManaCost: opt.Val(cost.Mana{
				cost.O(6),
				cost.B,
			}),
			Colors:    []color.Color{color.Black},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Elemental},
			Power:     opt.Val(game.PT{Value: 3}),
			Toughness: opt.Val(game.PT{Value: 3}),
			TriggeredAbilities: []game.TriggeredAbility{
				game.TriggeredAbility{
					Text: "When this creature enters, target player discards two cards.",
					Trigger: game.TriggerCondition{
						Type: game.TriggerWhen,
						Pattern: game.TriggerPattern{
							Event:           game.EventPermanentEnteredBattlefield,
							Source:          game.TriggerSourceSelf,
							StackObjectKind: game.StackSpell,
							CounterKind:     counter.PlusOnePlusOne,
						},
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
										Primitive: game.Discard{
											Amount: game.Fixed(2),
											Player: game.TargetPlayerReference(0),
										},
									},
								},
							},
						},
						MinModes: 1,
						MaxModes: 1,
					},
				},
				game.EvokeSacrificeTriggeredAbility(),
			},
			AlternativeCosts: []cost.Alternative{
				cost.Alternative{
					Label:    "Evoke",
					ManaCost: opt.Val(cost.Mana{cost.O(3), cost.B}),
					Mechanic: cost.AlternativeMechanicEvoke,
				},
			},
			OracleText: `
			When this creature enters, target player discards two cards.
			Evoke {3}{B} (You may cast this spell for its evoke cost. If you do, it's sacrificed when it enters.)
		`,
		},
	}
}
