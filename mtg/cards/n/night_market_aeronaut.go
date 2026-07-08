package n

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// NightMarketAeronaut is the card definition for Night Market Aeronaut.
//
// Type: Creature — Aetherborn Warrior
// Cost: {3}{B}
//
// Oracle text:
//
//	Flying
//	Revolt — This creature enters with a +1/+1 counter on it if a permanent left the battlefield under your control this turn.
var NightMarketAeronaut = func() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Black),
		CardFace: game.CardFace{
			Name: "Night Market Aeronaut",
			ManaCost: opt.Val(cost.Mana{
				cost.O(3),
				cost.B,
			}),
			Colors:    []color.Color{color.Black},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Aetherborn, types.Warrior},
			Power:     opt.Val(game.PT{Value: 2}),
			Toughness: opt.Val(game.PT{Value: 2}),
			StaticAbilities: []game.StaticAbility{
				game.FlyingStaticBody,
			},
			ReplacementAbilities: []game.ReplacementAbility{
				game.EntersWithCountersIfReplacement("Revolt — This creature enters with a +1/+1 counter on it if a permanent left the battlefield under your control this turn.", &game.Condition{
					EventHistory: opt.Val(game.EventHistoryCondition{Pattern: game.TriggerPattern{
						Event:         game.EventZoneChanged,
						Controller:    game.TriggerControllerYou,
						MatchFromZone: true,
						FromZone:      zone.Battlefield,
					}, Window: game.EventHistoryCurrentTurn}),
				}, game.CounterPlacement{Kind: counter.PlusOnePlusOne, Amount: 1}),
			},
			OracleText: `
			Flying
			Revolt — This creature enters with a +1/+1 counter on it if a permanent left the battlefield under your control this turn.
		`,
		},
	}
}
