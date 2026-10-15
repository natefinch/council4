package a

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// ArterialFlow is the card definition for Arterial Flow.
//
// Type: Sorcery
// Cost: {1}{B}{B}
//
// Oracle text:
//
//	Each opponent discards two cards. If you control a Vampire, each opponent loses 2 life and you gain 2 life.
var ArterialFlow = newArterialFlow

func newArterialFlow() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Black),
		CardFace: game.CardFace{
			Name: "Arterial Flow",
			ManaCost: opt.Val(cost.Mana{
				cost.O(1),
				cost.B,
				cost.B,
			}),
			Colors: []color.Color{color.Black},
			Types:  []types.Card{types.Sorcery},
			SpellAbility: opt.Val(game.Mode{
				Sequence: []game.Instruction{
					{
						Primitive: game.Discard{
							Amount:      game.Fixed(2),
							PlayerGroup: game.OpponentsReference(),
						},
					},
					{
						Primitive: game.LoseLife{
							Amount:      game.Fixed(2),
							PlayerGroup: game.OpponentsReference(),
						},
						Condition: opt.Val(game.EffectCondition{
							Condition: opt.Val(game.Condition{
								ControlsMatching: opt.Val(game.SelectionCount{
									Selection: game.Selection{SubtypesAny: []types.Sub{types.Sub("Vampire")}},
								}),
							}),
						}),
						PublishCondition: "condition-0",
					},
					{
						Primitive: game.GainLife{
							Amount: game.Fixed(2),
							Player: game.ControllerReference(),
						},
						ConditionGate: "condition-0",
					},
				},
			}.Ability()),
			OracleText: `
			Each opponent discards two cards. If you control a Vampire, each opponent loses 2 life and you gain 2 life.
		`,
		},
	}
}
