package f

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// FirestormHellkite is the card definition for Firestorm Hellkite.
//
// Type: Creature — Dragon
// Cost: {4}{U}{R}
//
// Oracle text:
//
//	Flying, trample
//	Cumulative upkeep {U}{R} (At the beginning of your upkeep, put an age counter on this permanent, then sacrifice it unless you pay its upkeep cost for each age counter on it.)
var FirestormHellkite = newFirestormHellkite

func newFirestormHellkite() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue, color.Red),
		CardFace: game.CardFace{
			Name: "Firestorm Hellkite",
			ManaCost: opt.Val(cost.Mana{
				cost.O(4),
				cost.U,
				cost.R,
			}),
			Colors:    []color.Color{color.Red, color.Blue},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Dragon},
			Power:     opt.Val(game.PT{Value: 6}),
			Toughness: opt.Val(game.PT{Value: 6}),
			StaticAbilities: []game.StaticAbility{
				game.FlyingStaticBody,
				game.TrampleStaticBody,
			},
			TriggeredAbilities: []game.TriggeredAbility{
				game.CumulativeUpkeepTriggeredAbility(cost.Mana{cost.U, cost.R}),
			},
			OracleText: `
			Flying, trample
			Cumulative upkeep {U}{R} (At the beginning of your upkeep, put an age counter on this permanent, then sacrifice it unless you pay its upkeep cost for each age counter on it.)
		`,
		},
	}
}
