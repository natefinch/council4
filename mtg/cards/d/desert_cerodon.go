package d

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// DesertCerodon is the card definition for Desert Cerodon.
//
// Type: Creature — Beast
// Cost: {5}{R}
//
// Oracle text:
//
//	Cycling {R} ({R}, Discard this card: Draw a card.)
var DesertCerodon = newDesertCerodon

func newDesertCerodon() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Red),
		CardFace: game.CardFace{
			Name: "Desert Cerodon",
			ManaCost: opt.Val(cost.Mana{
				cost.O(5),
				cost.R,
			}),
			Colors:    []color.Color{color.Red},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Beast},
			Power:     opt.Val(game.PT{Value: 6}),
			Toughness: opt.Val(game.PT{Value: 4}),
			ActivatedAbilities: []game.ActivatedAbility{
				game.CyclingActivatedAbility(cost.Mana{cost.R}),
			},
			OracleText: `
			Cycling {R} ({R}, Discard this card: Draw a card.)
		`,
		},
	}
}
