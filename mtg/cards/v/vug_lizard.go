package v

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// VugLizard is the card definition for Vug Lizard.
//
// Type: Creature — Lizard
// Cost: {1}{R}{R}
//
// Oracle text:
//
//	Mountainwalk (This creature can't be blocked as long as defending player controls a Mountain.)
//	Echo {1}{R}{R} (At the beginning of your upkeep, if this came under your control since the beginning of your last upkeep, sacrifice it unless you pay its echo cost.)
var VugLizard = newVugLizard

func newVugLizard() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Red),
		CardFace: game.CardFace{
			Name: "Vug Lizard",
			ManaCost: opt.Val(cost.Mana{
				cost.O(1),
				cost.R,
				cost.R,
			}),
			Colors:    []color.Color{color.Red},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Lizard},
			Power:     opt.Val(game.PT{Value: 3}),
			Toughness: opt.Val(game.PT{Value: 4}),
			StaticAbilities: []game.StaticAbility{
				game.MountainwalkStaticBody,
			},
			TriggeredAbilities: []game.TriggeredAbility{
				game.EchoTriggeredAbility(cost.Mana{cost.O(1), cost.R, cost.R}),
			},
			OracleText: `
			Mountainwalk (This creature can't be blocked as long as defending player controls a Mountain.)
			Echo {1}{R}{R} (At the beginning of your upkeep, if this came under your control since the beginning of your last upkeep, sacrifice it unless you pay its echo cost.)
		`,
		},
	}
}
