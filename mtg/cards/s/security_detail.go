package s

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// SecurityDetail is the card definition for Security Detail.
//
// Type: Enchantment
// Cost: {3}{W}
//
// Oracle text:
//
//	{W}{W}: Create a 1/1 white Soldier creature token. Activate only if you control no creatures and only once each turn.
var SecurityDetail = newSecurityDetail()

func newSecurityDetail() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White),
		CardFace: game.CardFace{
			Name: "Security Detail",
			ManaCost: opt.Val(cost.Mana{
				cost.O(3),
				cost.W,
			}),
			Colors: []color.Color{color.White},
			Types:  []types.Card{types.Enchantment},
			ActivatedAbilities: []game.ActivatedAbility{
				game.ActivatedAbility{
					Text:           "{W}{W}: Create a 1/1 white Soldier creature token. Activate only if you control no creatures and only once each turn.",
					ManaCost:       opt.Val(cost.Mana{cost.W, cost.W}),
					ZoneOfFunction: zone.Battlefield,
					Timing:         game.OncePerTurn,
					ActivationCondition: opt.Val(game.Condition{
						Negate: true,
						ControlsMatching: opt.Val(game.SelectionCount{
							Selection: game.Selection{RequiredTypes: []types.Card{types.Creature}},
							MinCount:  1,
						}),
					}),
					Content: game.Mode{
						Sequence: []game.Instruction{
							{
								Primitive: game.CreateToken{
									Amount: game.Fixed(1),
									Source: game.TokenDef(securityDetailToken),
								},
							},
						},
					}.Ability(),
				},
			},
			OracleText: `
			{W}{W}: Create a 1/1 white Soldier creature token. Activate only if you control no creatures and only once each turn.
		`,
		},
	}
}

var securityDetailToken = newSecurityDetailToken()

func newSecurityDetailToken() *game.CardDef {
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
