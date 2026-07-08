package l

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// LifecraftCavalry is the card definition for Lifecraft Cavalry.
//
// Type: Creature — Elf Warrior
// Cost: {4}{G}
//
// Oracle text:
//
//	Trample
//	Revolt — This creature enters with two +1/+1 counters on it if a permanent left the battlefield under your control this turn.
var LifecraftCavalry = func() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Green),
		CardFace: game.CardFace{
			Name: "Lifecraft Cavalry",
			ManaCost: opt.Val(cost.Mana{
				cost.O(4),
				cost.G,
			}),
			Colors:    []color.Color{color.Green},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Elf, types.Warrior},
			Power:     opt.Val(game.PT{Value: 4}),
			Toughness: opt.Val(game.PT{Value: 4}),
			StaticAbilities: []game.StaticAbility{
				game.TrampleStaticBody,
			},
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
			Trample
			Revolt — This creature enters with two +1/+1 counters on it if a permanent left the battlefield under your control this turn.
		`,
		},
	}
}
