package d

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// DeployToTheFront is the card definition for Deploy to the Front.
//
// Type: Sorcery
// Cost: {5}{W}{W}
//
// Oracle text:
//
//	Create X 1/1 white Soldier creature tokens, where X is the number of creatures on the battlefield.
var DeployToTheFront = newDeployToTheFront

func newDeployToTheFront() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White),
		CardFace: game.CardFace{
			Name: "Deploy to the Front",
			ManaCost: opt.Val(cost.Mana{
				cost.O(5),
				cost.W,
				cost.W,
			}),
			Colors: []color.Color{color.White},
			Types:  []types.Card{types.Sorcery},
			SpellAbility: opt.Val(game.Mode{
				Sequence: []game.Instruction{
					{
						Primitive: game.CreateToken{
							Amount: game.Dynamic(game.DynamicAmount{
								Kind:       game.DynamicAmountCountSelector,
								Multiplier: 1,
								Group:      game.BattlefieldGroup(game.Selection{RequiredTypes: []types.Card{types.Creature}}),
							}),
							Source: game.TokenDef(deployToTheFrontToken),
						},
					},
				},
			}.Ability()),
			OracleText: `
			Create X 1/1 white Soldier creature tokens, where X is the number of creatures on the battlefield.
		`,
		},
	}
}

var deployToTheFrontToken = newDeployToTheFrontToken()

func newDeployToTheFrontToken() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name:      "Soldier",
			Colors:    []color.Color{color.White},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Soldier},
			Power:     opt.Val(game.PT{Value: 1}),
			Toughness: opt.Val(game.PT{Value: 1}),
		},
	}
}
