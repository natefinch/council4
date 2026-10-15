package i

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// InfernalContract is the card definition for Infernal Contract.
//
// Type: Sorcery
// Cost: {B}{B}{B}
//
// Oracle text:
//
//	Draw four cards. You lose half your life, rounded up.
var InfernalContract = newInfernalContract

func newInfernalContract() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Black),
		CardFace: game.CardFace{
			Name: "Infernal Contract",
			ManaCost: opt.Val(cost.Mana{
				cost.B,
				cost.B,
				cost.B,
			}),
			Colors: []color.Color{color.Black},
			Types:  []types.Card{types.Sorcery},
			SpellAbility: opt.Val(game.Mode{
				Sequence: []game.Instruction{
					{
						Primitive: game.Draw{
							Amount: game.Fixed(4),
							Player: game.ControllerReference(),
						},
					},
					{
						Primitive: game.LoseLife{
							Amount: game.Dynamic(game.DynamicAmount{
								Kind:    game.DynamicAmountPlayerLife,
								Divisor: 2,
								RoundUp: true,
								Player:  func() *game.PlayerReference { ref := game.ControllerReference(); return &ref }(),
							}),
							Player: game.ControllerReference(),
						},
					},
				},
			}.Ability()),
			OracleText: `
			Draw four cards. You lose half your life, rounded up.
		`,
		},
	}
}
