package t

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// TectonicFiend is the card definition for Tectonic Fiend.
//
// Type: Creature — Elemental
// Cost: {4}{R}{R}
//
// Oracle text:
//
//	Echo {4}{R}{R} (At the beginning of your upkeep, if this came under your control since the beginning of your last upkeep, sacrifice it unless you pay its echo cost.)
//	This creature attacks each combat if able.
var TectonicFiend = newTectonicFiend

func newTectonicFiend() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Red),
		CardFace: game.CardFace{
			Name: "Tectonic Fiend",
			ManaCost: opt.Val(cost.Mana{
				cost.O(4),
				cost.R,
				cost.R,
			}),
			Colors:    []color.Color{color.Red},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Elemental},
			Power:     opt.Val(game.PT{Value: 7}),
			Toughness: opt.Val(game.PT{Value: 7}),
			StaticAbilities: []game.StaticAbility{
				game.MustAttackStaticBody,
			},
			TriggeredAbilities: []game.TriggeredAbility{
				game.EchoTriggeredAbility(cost.Mana{cost.O(4), cost.R, cost.R}),
			},
			OracleText: `
			Echo {4}{R}{R} (At the beginning of your upkeep, if this came under your control since the beginning of your last upkeep, sacrifice it unless you pay its echo cost.)
			This creature attacks each combat if able.
		`,
		},
	}
}
