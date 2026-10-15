package s

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// SphinxSInsight is the card definition for Sphinx's Insight.
//
// Type: Instant
// Cost: {2}{W}{U}
//
// Oracle text:
//
//	Draw two cards.
//	Addendum — If you cast this spell during your main phase, you gain 2 life.
var SphinxSInsight = newSphinxSInsight

func newSphinxSInsight() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White, color.Blue),
		CardFace: game.CardFace{
			Name: "Sphinx's Insight",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
				cost.W,
				cost.U,
			}),
			Colors: []color.Color{color.Blue, color.White},
			Types:  []types.Card{types.Instant},
			SpellAbility: opt.Val(game.Mode{
				Sequence: []game.Instruction{
					{
						Primitive: game.Draw{
							Amount: game.Fixed(2),
							Player: game.ControllerReference(),
						},
					},
					{
						Primitive: game.GainLife{
							Amount: game.Fixed(2),
							Player: game.ControllerReference(),
						},
						Condition: opt.Val(game.EffectCondition{
							Condition: opt.Val(game.Condition{
								CastDuringControllerMainPhase: true,
							}),
						}),
					},
				},
			}.Ability()),
			OracleText: `
			Draw two cards.
			Addendum — If you cast this spell during your main phase, you gain 2 life.
		`,
		},
	}
}
