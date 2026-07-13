package g

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// GoblinWarBuggy is the card definition for Goblin War Buggy.
//
// Type: Creature — Goblin
// Cost: {1}{R}
//
// Oracle text:
//
//	Haste
//	Echo {1}{R} (At the beginning of your upkeep, if this came under your control since the beginning of your last upkeep, sacrifice it unless you pay its echo cost.)
var GoblinWarBuggy = newGoblinWarBuggy

func newGoblinWarBuggy() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Red),
		CardFace: game.CardFace{
			Name: "Goblin War Buggy",
			ManaCost: opt.Val(cost.Mana{
				cost.O(1),
				cost.R,
			}),
			Colors:    []color.Color{color.Red},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Goblin},
			Power:     opt.Val(game.PT{Value: 2}),
			Toughness: opt.Val(game.PT{Value: 2}),
			StaticAbilities: []game.StaticAbility{
				game.HasteStaticBody,
			},
			TriggeredAbilities: []game.TriggeredAbility{
				game.EchoTriggeredAbility(cost.Mana{cost.O(1), cost.R}),
			},
			OracleText: `
			Haste
			Echo {1}{R} (At the beginning of your upkeep, if this came under your control since the beginning of your last upkeep, sacrifice it unless you pay its echo cost.)
		`,
		},
	}
}
