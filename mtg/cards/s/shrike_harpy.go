package s

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// ShrikeHarpy is the card definition for Shrike Harpy.
//
// Type: Creature — Harpy
// Cost: {3}{B}{B}
//
// Oracle text:
//
//	Flying
//	Tribute 2 (As this creature enters, an opponent of your choice may put two +1/+1 counters on it.)
//	When this creature enters, if tribute wasn't paid, target opponent sacrifices a creature of their choice.
var ShrikeHarpy = newShrikeHarpy

func newShrikeHarpy() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Black),
		CardFace: game.CardFace{
			Name: "Shrike Harpy",
			ManaCost: opt.Val(cost.Mana{
				cost.O(3),
				cost.B,
				cost.B,
			}),
			Colors:    []color.Color{color.Black},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Harpy},
			Power:     opt.Val(game.PT{Value: 2}),
			Toughness: opt.Val(game.PT{Value: 2}),
			StaticAbilities: []game.StaticAbility{
				game.FlyingStaticBody,
			},
			TriggeredAbilities: []game.TriggeredAbility{
				game.TriggeredAbility{
					Text: "When this creature enters, if tribute wasn't paid, target opponent sacrifices a creature of their choice.",
					Trigger: game.TriggerCondition{
						Type: game.TriggerWhen,
						Pattern: game.TriggerPattern{
							Event:           game.EventPermanentEnteredBattlefield,
							Source:          game.TriggerSourceSelf,
							StackObjectKind: game.StackSpell,
							CounterKind:     counter.PlusOnePlusOne,
						},
						InterveningIf: "if tribute wasn't paid",
						InterveningCondition: opt.Val(game.Condition{
							Text:                 "if tribute wasn't paid",
							SourceCounterKind:    counter.PlusOnePlusOne,
							SourceTributeNotPaid: true,
						}),
					},
					Content: game.AbilityContent{
						Modes: []game.Mode{
							game.Mode{
								Targets: []game.TargetSpec{
									game.TargetSpec{
										MinTargets: 1,
										MaxTargets: 1,
										Constraint: "target opponent",
										Allow:      game.TargetAllowPlayer,
										Selection:  opt.Val(game.Selection{Player: game.PlayerOpponent}),
									},
								},
								Sequence: []game.Instruction{
									game.Instruction{
										Primitive: game.SacrificePermanents{
											Player:    game.TargetPlayerReference(0),
											Amount:    game.Fixed(1),
											Selection: game.Selection{RequiredTypes: []types.Card{types.Creature}},
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
			ReplacementAbilities: []game.ReplacementAbility{
				game.TributeReplacement("As this creature enters, an opponent of your choice may put 2 +1/+1 counters on it.", 2),
			},
			OracleText: `
			Flying
			Tribute 2 (As this creature enters, an opponent of your choice may put two +1/+1 counters on it.)
			When this creature enters, if tribute wasn't paid, target opponent sacrifices a creature of their choice.
		`,
		},
	}
}
