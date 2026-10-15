package w

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// WinterSoldierBuckyBarnes is the card definition for Winter Soldier, Bucky Barnes.
//
// Type: Legendary Creature — Human Soldier Hero
// Cost: {W}
//
// Oracle text:
//
//	Winter Soldier enters tapped.
var WinterSoldierBuckyBarnes = newWinterSoldierBuckyBarnes

func newWinterSoldierBuckyBarnes() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White),
		CardFace: game.CardFace{
			Name: "Winter Soldier, Bucky Barnes",
			ManaCost: opt.Val(cost.Mana{
				cost.W,
			}),
			Colors:     []color.Color{color.White},
			Supertypes: []types.Super{types.Legendary},
			Types:      []types.Card{types.Creature},
			Subtypes:   []types.Sub{types.Human, types.Soldier, types.Hero},
			Power:      opt.Val(game.PT{Value: 2}),
			Toughness:  opt.Val(game.PT{Value: 2}),
			ReplacementAbilities: []game.ReplacementAbility{
				game.EntersTappedReplacement("Winter Soldier enters tapped."),
			},
			OracleText: `
			Winter Soldier enters tapped.
		`,
		},
	}
}
