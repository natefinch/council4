package v

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// ViashinoOutrider is the card definition for Viashino Outrider.
//
// Type: Creature — Lizard
// Cost: {2}{R}
//
// Oracle text:
//
//	Echo {2}{R} (At the beginning of your upkeep, if this came under your control since the beginning of your last upkeep, sacrifice it unless you pay its echo cost.)
var ViashinoOutrider = newViashinoOutrider

func newViashinoOutrider() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Red),
		CardFace: game.CardFace{
			Name: "Viashino Outrider",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
				cost.R,
			}),
			Colors:    []color.Color{color.Red},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Lizard},
			Power:     opt.Val(game.PT{Value: 4}),
			Toughness: opt.Val(game.PT{Value: 3}),
			TriggeredAbilities: []game.TriggeredAbility{
				game.EchoTriggeredAbility(cost.Mana{cost.O(2), cost.R}),
			},
			OracleText: `
			Echo {2}{R} (At the beginning of your upkeep, if this came under your control since the beginning of your last upkeep, sacrifice it unless you pay its echo cost.)
		`,
		},
	}
}
