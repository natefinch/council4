package b

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// BringerOfTheGreenDawn is the card definition for Bringer of the Green Dawn.
//
// Type: Creature — Bringer
// Cost: {7}{G}{G}
//
// Oracle text:
//
//	You may pay {W}{U}{B}{R}{G} rather than pay this spell's mana cost.
//	Trample
//	At the beginning of your upkeep, you may create a 3/3 green Beast creature token.
var BringerOfTheGreenDawn = newBringerOfTheGreenDawn

func newBringerOfTheGreenDawn() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White, color.Blue, color.Black, color.Red, color.Green),
		CardFace: game.CardFace{
			Name: "Bringer of the Green Dawn",
			ManaCost: opt.Val(cost.Mana{
				cost.O(7),
				cost.G,
				cost.G,
			}),
			Colors:    []color.Color{color.Green},
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
								Primitive: game.CreateToken{
									Amount: game.Fixed(1),
									Source: game.TokenDef(bringerOfTheGreenDawnToken),
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
			At the beginning of your upkeep, you may create a 3/3 green Beast creature token.
		`,
		},
	}
}

var bringerOfTheGreenDawnToken = newBringerOfTheGreenDawnToken()

func newBringerOfTheGreenDawnToken() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name:      "Beast",
			Colors:    []color.Color{color.Green},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Beast},
			Power:     opt.Val(game.PT{Value: 3}),
			Toughness: opt.Val(game.PT{Value: 3}),
		},
	}
}
