package y

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// YokedPlowbeast is the card definition for Yoked Plowbeast.
//
// Type: Creature — Beast
// Cost: {5}{W}{W}
//
// Oracle text:
//
//	Cycling {2} ({2}, Discard this card: Draw a card.)
var YokedPlowbeast = newYokedPlowbeast

func newYokedPlowbeast() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White),
		CardFace: game.CardFace{
			Name: "Yoked Plowbeast",
			ManaCost: opt.Val(cost.Mana{
				cost.O(5),
				cost.W,
				cost.W,
			}),
			Colors:    []color.Color{color.White},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Beast},
			Power:     opt.Val(game.PT{Value: 5}),
			Toughness: opt.Val(game.PT{Value: 5}),
			ActivatedAbilities: []game.ActivatedAbility{
				game.CyclingActivatedAbility(cost.Mana{cost.O(2)}),
			},
			OracleText: `
			Cycling {2} ({2}, Discard this card: Draw a card.)
		`,
		},
	}
}
