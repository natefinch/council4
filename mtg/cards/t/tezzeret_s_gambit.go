package t

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// TezzeretSGambit is the card definition for Tezzeret's Gambit.
//
// Type: Sorcery
// Cost: {3}{U/P}
//
// Oracle text:
//
//	({U/P} can be paid with either {U} or 2 life.)
//	Draw two cards, then proliferate. (Choose any number of permanents and/or players, then give each another counter of each kind already there.)
var TezzeretSGambit = newTezzeretSGambit

func newTezzeretSGambit() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue),
		CardFace: game.CardFace{
			Name: "Tezzeret's Gambit",
			ManaCost: opt.Val(cost.Mana{
				cost.O(3),
				cost.PhyrexianMana(mana.U),
			}),
			Colors: []color.Color{color.Blue},
			Types:  []types.Card{types.Sorcery},
			SpellAbility: opt.Val(game.Mode{
				Sequence: []game.Instruction{
					{
						Primitive: game.Draw{
							Amount: game.Fixed(2),
							Player: game.ControllerReference(),
						},
					},
					{
						Primitive: game.Proliferate{
							Amount: game.Fixed(1),
						},
					},
				},
			}.Ability()),
			OracleText: `
			({U/P} can be paid with either {U} or 2 life.)
			Draw two cards, then proliferate. (Choose any number of permanents and/or players, then give each another counter of each kind already there.)
		`,
		},
	}
}
