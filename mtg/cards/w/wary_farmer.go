package w

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// WaryFarmer is the card definition for Wary Farmer.
//
// Type: Creature — Kithkin Citizen
// Cost: {1}{G/W}{G/W}
//
// Oracle text:
//
//	At the beginning of your end step, if another creature entered the battlefield under your control this turn, surveil 1. (Look at the top card of your library. You may put it into your graveyard.)
var WaryFarmer = newWaryFarmer

func newWaryFarmer() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White, color.Green),
		CardFace: game.CardFace{
			Name: "Wary Farmer",
			ManaCost: opt.Val(cost.Mana{
				cost.O(1),
				cost.HybridMana(mana.G, mana.W),
				cost.HybridMana(mana.G, mana.W),
			}),
			Colors:    []color.Color{color.Green, color.White},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Kithkin, types.Citizen},
			Power:     opt.Val(game.PT{Value: 3}),
			Toughness: opt.Val(game.PT{Value: 3}),
			TriggeredAbilities: []game.TriggeredAbility{
				game.TriggeredAbility{
					Trigger: game.TriggerCondition{
						Type: game.TriggerAt,
						Pattern: game.TriggerPattern{
							Event:      game.EventBeginningOfStep,
							Controller: game.TriggerControllerYou,
							Step:       game.StepEnd,
						},
						InterveningIf: "if another creature entered the battlefield under your control this turn",
						InterveningCondition: opt.Val(game.Condition{
							EventHistory: opt.Val(game.EventHistoryCondition{Pattern: game.TriggerPattern{
								Event:            game.EventPermanentEnteredBattlefield,
								Controller:       game.TriggerControllerYou,
								ExcludeSelf:      true,
								SubjectSelection: game.Selection{RequiredTypes: []types.Card{types.Creature}},
							}, Window: game.EventHistoryCurrentTurn}),
						}),
					},
					Content: game.Mode{
						Sequence: []game.Instruction{
							{
								Primitive: game.Surveil{
									Amount: game.Fixed(1),
									Player: game.ControllerReference(),
								},
							},
						},
					}.Ability(),
				},
			},
			OracleText: `
			At the beginning of your end step, if another creature entered the battlefield under your control this turn, surveil 1. (Look at the top card of your library. You may put it into your graveyard.)
		`,
		},
	}
}
