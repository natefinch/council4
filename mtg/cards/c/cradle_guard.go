package c

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// CradleGuard is the card definition for Cradle Guard.
//
// Type: Creature — Treefolk
// Cost: {1}{G}{G}
//
// Oracle text:
//
//	Trample
//	Echo {1}{G}{G} (At the beginning of your upkeep, if this came under your control since the beginning of your last upkeep, sacrifice it unless you pay its echo cost.)
var CradleGuard = newCradleGuard

func newCradleGuard() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Green),
		CardFace: game.CardFace{
			Name: "Cradle Guard",
			ManaCost: opt.Val(cost.Mana{
				cost.O(1),
				cost.G,
				cost.G,
			}),
			Colors:    []color.Color{color.Green},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Treefolk},
			Power:     opt.Val(game.PT{Value: 4}),
			Toughness: opt.Val(game.PT{Value: 4}),
			StaticAbilities: []game.StaticAbility{
				game.TrampleStaticBody,
			},
			TriggeredAbilities: []game.TriggeredAbility{
				game.EchoTriggeredAbility(cost.Mana{cost.O(1), cost.G, cost.G}),
			},
			OracleText: `
			Trample
			Echo {1}{G}{G} (At the beginning of your upkeep, if this came under your control since the beginning of your last upkeep, sacrifice it unless you pay its echo cost.)
		`,
		},
	}
}
