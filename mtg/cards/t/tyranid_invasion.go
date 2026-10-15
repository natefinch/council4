package t

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// TyranidInvasion is the card definition for Tyranid Invasion.
//
// Type: Sorcery
// Cost: {3}{G}
//
// Oracle text:
//
//	Create a number of 3/3 green Tyranid Warrior creature tokens with trample equal to the number of opponents you have.
var TyranidInvasion = newTyranidInvasion

func newTyranidInvasion() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Green),
		CardFace: game.CardFace{
			Name: "Tyranid Invasion",
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
								Kind:       game.DynamicAmountOpponentCount,
								Multiplier: 1,
							}),
							Source: game.TokenDef(tyranidInvasionToken),
						},
					},
				},
			}.Ability()),
			OracleText: `
			Create a number of 3/3 green Tyranid Warrior creature tokens with trample equal to the number of opponents you have.
		`,
		},
	}
}

var tyranidInvasionToken = newTyranidInvasionToken()

func newTyranidInvasionToken() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name:      "Tyranid Warrior",
			Colors:    []color.Color{color.Green},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Tyranid, types.Warrior},
			Power:     opt.Val(game.PT{Value: 3}),
			Toughness: opt.Val(game.PT{Value: 3}),
			StaticAbilities: []game.StaticAbility{
				game.TrampleStaticBody,
			},
		},
	}
}
