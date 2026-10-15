package t

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// TheCrystalSChosen is the card definition for The Crystal's Chosen.
//
// Type: Sorcery
// Cost: {5}{W}{W}
//
// Oracle text:
//
//	Create four 1/1 colorless Hero creature tokens. Then put a +1/+1 counter on each creature you control.
var TheCrystalSChosen = newTheCrystalSChosen

func newTheCrystalSChosen() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White),
		CardFace: game.CardFace{
			Name: "The Crystal's Chosen",
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
							Amount: game.Fixed(4),
							Source: game.TokenDef(theCrystalSChosenToken),
						},
					},
					{
						Primitive: game.AddCounter{
							Amount:      game.Fixed(1),
							Group:       game.BattlefieldGroup(game.Selection{RequiredTypes: []types.Card{types.Creature}, Controller: game.ControllerYou}),
							CounterKind: counter.PlusOnePlusOne,
						},
					},
				},
			}.Ability()),
			OracleText: `
			Create four 1/1 colorless Hero creature tokens. Then put a +1/+1 counter on each creature you control.
		`,
		},
	}
}

var theCrystalSChosenToken = newTheCrystalSChosenToken()

func newTheCrystalSChosenToken() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name:      "Hero",
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Hero},
			Power:     opt.Val(game.PT{Value: 1}),
			Toughness: opt.Val(game.PT{Value: 1}),
		},
	}
}
