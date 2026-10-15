package f

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// FungalSprouting is the card definition for Fungal Sprouting.
//
// Type: Sorcery
// Cost: {3}{G}
//
// Oracle text:
//
//	Create X 1/1 green Saproling creature tokens, where X is the greatest power among creatures you control.
var FungalSprouting = newFungalSprouting

func newFungalSprouting() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Green),
		CardFace: game.CardFace{
			Name: "Fungal Sprouting",
			ManaCost: opt.Val(cost.Mana{
				cost.O(3),
				cost.G,
			}),
			Colors: []color.Color{color.Green},
			Types:  []types.Card{types.Sorcery},
			SpellAbility: opt.Val(game.Mode{
				Sequence: []game.Instruction{
					{
						Primitive: game.CreateToken{
							Amount: game.Dynamic(game.DynamicAmount{
								Kind:       game.DynamicAmountGreatestPowerInGroup,
								Multiplier: 1,
								Group:      game.BattlefieldGroup(game.Selection{RequiredTypes: []types.Card{types.Creature}, Controller: game.ControllerYou}),
							}),
							Source: game.TokenDef(fungalSproutingToken),
						},
					},
				},
			}.Ability()),
			OracleText: `
			Create X 1/1 green Saproling creature tokens, where X is the greatest power among creatures you control.
		`,
		},
	}
}

var fungalSproutingToken = newFungalSproutingToken()

func newFungalSproutingToken() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name:      "Saproling",
			Colors:    []color.Color{color.Green},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Saproling},
			Power:     opt.Val(game.PT{Value: 1}),
			Toughness: opt.Val(game.PT{Value: 1}),
		},
	}
}
