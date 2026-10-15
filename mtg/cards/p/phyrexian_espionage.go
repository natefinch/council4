package p

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// PhyrexianEspionage is the card definition for Phyrexian Espionage.
//
// Type: Sorcery
// Cost: {2}{U}
//
// Oracle text:
//
//	Kicker {1}{B} (You may pay an additional {1}{B} as you cast this spell.)
//	Draw two cards. If this spell was kicked, each opponent discards a card.
var PhyrexianEspionage = newPhyrexianEspionage

func newPhyrexianEspionage() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue, color.Black),
		CardFace: game.CardFace{
			Name: "Phyrexian Espionage",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
				cost.U,
			}),
			Colors: []color.Color{color.Blue},
			Types:  []types.Card{types.Sorcery},
			StaticAbilities: []game.StaticAbility{
				game.StaticAbility{
					Text: "Kicker {1}{B}",
					KeywordAbilities: []game.KeywordAbility{
						game.KickerKeyword{
							Cost: cost.Mana{
								cost.O(1),
								cost.B,
							},
						},
					},
				},
			},
			SpellAbility: opt.Val(game.Mode{
				Sequence: []game.Instruction{
					{
						Primitive: game.Draw{
							Amount: game.Fixed(2),
							Player: game.ControllerReference(),
						},
					},
					{
						Primitive: game.Discard{
							Amount:      game.Fixed(1),
							PlayerGroup: game.OpponentsReference(),
						},
						Condition: opt.Val(game.EffectCondition{
							Condition: opt.Val(game.Condition{
								SpellWasKicked: true,
							}),
						}),
					},
				},
			}.Ability()),
			OracleText: `
			Kicker {1}{B} (You may pay an additional {1}{B} as you cast this spell.)
			Draw two cards. If this spell was kicked, each opponent discards a card.
		`,
		},
	}
}
