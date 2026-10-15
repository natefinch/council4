package a

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// AngelOfTheGodPharaoh is the card definition for Angel of the God-Pharaoh.
//
// Type: Creature — Angel
// Cost: {4}{W}{W}
//
// Oracle text:
//
//	Flying
//	Cycling {2} ({2}, Discard this card: Draw a card.)
var AngelOfTheGodPharaoh = newAngelOfTheGodPharaoh

func newAngelOfTheGodPharaoh() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White),
		CardFace: game.CardFace{
			Name: "Angel of the God-Pharaoh",
			ManaCost: opt.Val(cost.Mana{
				cost.O(4),
				cost.W,
				cost.W,
			}),
			Colors:    []color.Color{color.White},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Angel},
			Power:     opt.Val(game.PT{Value: 4}),
			Toughness: opt.Val(game.PT{Value: 4}),
			StaticAbilities: []game.StaticAbility{
				game.FlyingStaticBody,
			},
			ActivatedAbilities: []game.ActivatedAbility{
				game.CyclingActivatedAbility(cost.Mana{cost.O(2)}),
			},
			OracleText: `
			Flying
			Cycling {2} ({2}, Discard this card: Draw a card.)
		`,
		},
	}
}
