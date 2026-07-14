package s

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// SilverflameRitual is the card definition for Silverflame Ritual.
//
// Type: Sorcery
// Cost: {3}{W}
//
// Oracle text:
//
//	Put a +1/+1 counter on each creature you control.
//	Adamant — If at least three white mana was spent to cast this spell, creatures you control gain vigilance until end of turn.
var SilverflameRitual = newSilverflameRitual

func newSilverflameRitual() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White),
		CardFace: game.CardFace{
			Name: "Silverflame Ritual",
			ManaCost: opt.Val(cost.Mana{
				cost.O(3),
				cost.W,
			}),
			Colors: []color.Color{color.White},
			Types:  []types.Card{types.Sorcery},
			SpellAbility: opt.Val(game.Mode{
				Sequence: []game.Instruction{
					{
						Primitive: game.AddCounter{
							Amount:      game.Fixed(1),
							Group:       game.BattlefieldGroup(game.Selection{RequiredTypes: []types.Card{types.Creature}, Controller: game.ControllerYou}),
							CounterKind: counter.PlusOnePlusOne,
						},
					},
					{
						Primitive: game.ApplyContinuous{
							ContinuousEffects: []game.ContinuousEffect{
								game.ContinuousEffect{
									Layer: game.LayerAbility,
									Group: game.BattlefieldGroup(game.Selection{RequiredTypes: []types.Card{types.Creature}, Controller: game.ControllerYou}),
									AddKeywords: []game.Keyword{
										game.Vigilance,
									},
								},
							},
							Duration: game.DurationUntilEndOfTurn,
						},
						Condition: opt.Val(game.EffectCondition{
							Condition: opt.Val(game.Condition{
								SpellColorManaSpent: game.ColorManaSpendThreshold{Color: color.White, Count: 3},
							}),
						}),
					},
				},
			}.Ability()),
			OracleText: `
			Put a +1/+1 counter on each creature you control.
			Adamant — If at least three white mana was spent to cast this spell, creatures you control gain vigilance until end of turn.
		`,
		},
	}
}
