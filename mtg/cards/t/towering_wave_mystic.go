package t

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// ToweringWaveMystic is the card definition for Towering-Wave Mystic.
//
// Type: Creature — Merfolk Wizard
// Cost: {1}{U}
//
// Oracle text:
//
//	Whenever this creature deals damage, target player mills that many cards.
var ToweringWaveMystic = newToweringWaveMystic()

func newToweringWaveMystic() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue),
		CardFace: game.CardFace{
			Name: "Towering-Wave Mystic",
			ManaCost: opt.Val(cost.Mana{
				cost.O(1),
				cost.U,
			}),
			Colors:    []color.Color{color.Blue},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Merfolk, types.Wizard},
			Power:     opt.Val(game.PT{Value: 2}),
			Toughness: opt.Val(game.PT{Value: 1}),
			TriggeredAbilities: []game.TriggeredAbility{
				game.TriggeredAbility{
					Trigger: game.TriggerCondition{
						Type: game.TriggerWhenever,
						Pattern: game.TriggerPattern{
							Event:   game.EventDamageDealt,
							Source:  game.TriggerSourceSelf,
							Subject: game.TriggerSubjectDamageSource,
						},
					},
					Content: game.Mode{
						Targets: []game.TargetSpec{
							game.TargetSpec{
								MinTargets: 1,
								MaxTargets: 1,
								Constraint: "target player",
								Allow:      game.TargetAllowPlayer,
							},
						},
						Sequence: []game.Instruction{
							{
								Primitive: game.Mill{
									Amount: game.Dynamic(game.DynamicAmount{
										Kind:       game.DynamicAmountEventDamage,
										Multiplier: 1,
									}),
									Player: game.TargetPlayerReference(0),
								},
							},
						},
					}.Ability(),
				},
			},
			OracleText: `
			Whenever this creature deals damage, target player mills that many cards.
		`,
		},
	}
}
