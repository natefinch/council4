package f

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// FlamecoreElemental is the card definition for Flamecore Elemental.
//
// Type: Creature — Elemental
// Cost: {2}{R}{R}
//
// Oracle text:
//
//	Echo {2}{R}{R} (At the beginning of your upkeep, if this came under your control since the beginning of your last upkeep, sacrifice it unless you pay its echo cost.)
var FlamecoreElemental = newFlamecoreElemental

func newFlamecoreElemental() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Red),
		CardFace: game.CardFace{
			Name: "Flamecore Elemental",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
				cost.R,
				cost.R,
			}),
			Colors:    []color.Color{color.Red},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Elemental},
			Power:     opt.Val(game.PT{Value: 5}),
			Toughness: opt.Val(game.PT{Value: 4}),
			TriggeredAbilities: []game.TriggeredAbility{
				game.EchoTriggeredAbility(cost.Mana{cost.O(2), cost.R, cost.R}),
			},
			OracleText: `
			Echo {2}{R}{R} (At the beginning of your upkeep, if this came under your control since the beginning of your last upkeep, sacrifice it unless you pay its echo cost.)
		`,
		},
	}
}
