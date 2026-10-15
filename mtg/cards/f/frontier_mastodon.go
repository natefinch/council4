package f

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/compare"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// FrontierMastodon is the card definition for Frontier Mastodon.
//
// Type: Creature — Elephant
// Cost: {2}{G}
//
// Oracle text:
//
//	Ferocious — This creature enters with a +1/+1 counter on it if you control a creature with power 4 or greater.
var FrontierMastodon = newFrontierMastodon

func newFrontierMastodon() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Green),
		CardFace: game.CardFace{
			Name: "Frontier Mastodon",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
				cost.G,
			}),
			Colors:    []color.Color{color.Green},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Elephant},
			Power:     opt.Val(game.PT{Value: 3}),
			Toughness: opt.Val(game.PT{Value: 2}),
			ReplacementAbilities: []game.ReplacementAbility{
				game.EntersWithCountersIfReplacement("Ferocious — This creature enters with a +1/+1 counter on it if you control a creature with power 4 or greater.", &game.Condition{
					ControlsMatching: opt.Val(game.SelectionCount{
						Selection: game.Selection{RequiredTypes: []types.Card{types.Creature}, Power: opt.Val(compare.Int{Op: compare.GreaterOrEqual, Value: 4})},
					}),
				}, game.CounterPlacement{Kind: counter.PlusOnePlusOne, Amount: 1}),
			},
			OracleText: `
			Ferocious — This creature enters with a +1/+1 counter on it if you control a creature with power 4 or greater.
		`,
		},
	}
}
