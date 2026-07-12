package b

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// BringerOfTheBlueDawn is the card definition for Bringer of the Blue Dawn.
//
// Type: Creature — Bringer
// Cost: {7}{U}{U}
//
// Oracle text:
//
//	You may pay {W}{U}{B}{R}{G} rather than pay this spell's mana cost.
//	Trample
//	At the beginning of your upkeep, you may draw two cards.
var BringerOfTheBlueDawn = newBringerOfTheBlueDawn

func newBringerOfTheBlueDawn() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White, color.Blue, color.Black, color.Red, color.Green),
		CardFace: game.CardFace{
			Name: "Bringer of the Blue Dawn",
			ManaCost: opt.Val(cost.Mana{
				cost.O(7),
				cost.U,
				cost.U,
			}),
			Colors:    []color.Color{color.Blue},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Bringer},
			Power:     opt.Val(game.PT{Value: 5}),
			Toughness: opt.Val(game.PT{Value: 5}),
			StaticAbilities: []game.StaticAbility{
				game.TrampleStaticBody,
			},
			TriggeredAbilities: []game.TriggeredAbility{
				game.TriggeredAbility{
					Trigger: game.TriggerCondition{
						Type: game.TriggerAt,
						Pattern: game.TriggerPattern{
							Event:      game.EventBeginningOfStep,
							Controller: game.TriggerControllerYou,
							Step:       game.StepUpkeep,
						},
					},
					Optional: true,
					Content: game.Mode{
						Sequence: []game.Instruction{
							{
								Primitive: game.Draw{
									Amount: game.Fixed(2),
									Player: game.ControllerReference(),
								},
							},
						},
					}.Ability(),
				},
			},
			AlternativeCosts: []cost.Alternative{
				cost.Alternative{
					Label:    "Pay {W}{U}{B}{R}{G}",
					ManaCost: opt.Val(cost.Mana{cost.W, cost.U, cost.B, cost.R, cost.G}),
				},
			},
			OracleText: `
			You may pay {W}{U}{B}{R}{G} rather than pay this spell's mana cost.
			Trample
			At the beginning of your upkeep, you may draw two cards.
		`,
		},
	}
}
