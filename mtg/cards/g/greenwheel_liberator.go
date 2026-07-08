package g

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// GreenwheelLiberator is the card definition for Greenwheel Liberator.
//
// Type: Creature — Elf Warrior
// Cost: {1}{G}
//
// Oracle text:
//
//	Revolt — This creature enters with two +1/+1 counters on it if a permanent left the battlefield under your control this turn.
var GreenwheelLiberator = func() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Green),
		CardFace: game.CardFace{
			Name: "Greenwheel Liberator",
			ManaCost: opt.Val(cost.Mana{
				cost.O(1),
				cost.G,
			}),
			Colors:    []color.Color{color.Green},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Elf, types.Warrior},
			Power:     opt.Val(game.PT{Value: 2}),
			Toughness: opt.Val(game.PT{Value: 1}),
			ReplacementAbilities: []game.ReplacementAbility{
				game.EntersWithCountersIfReplacement("Revolt — This creature enters with two +1/+1 counters on it if a permanent left the battlefield under your control this turn.", &game.Condition{
					EventHistory: opt.Val(game.EventHistoryCondition{Pattern: game.TriggerPattern{
						Event:         game.EventZoneChanged,
						Controller:    game.TriggerControllerYou,
						MatchFromZone: true,
						FromZone:      zone.Battlefield,
					}, Window: game.EventHistoryCurrentTurn}),
				}, game.CounterPlacement{Kind: counter.PlusOnePlusOne, Amount: 2}),
			},
			OracleText: `
			Revolt — This creature enters with two +1/+1 counters on it if a permanent left the battlefield under your control this turn.
		`,
		},
	}
}
