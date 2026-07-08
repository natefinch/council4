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

// NarnamRenegade is the card definition for Narnam Renegade.
//
// Type: Creature — Elf Warrior
// Cost: {G}
//
// Oracle text:
//
//	Deathtouch
//	Revolt — This creature enters with a +1/+1 counter on it if a permanent left the battlefield under your control this turn.
var NarnamRenegade = func() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Green),
		CardFace: game.CardFace{
			Name: "Narnam Renegade",
			ManaCost: opt.Val(cost.Mana{
				cost.G,
			}),
			Colors:    []color.Color{color.Green},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Elf, types.Warrior},
			Power:     opt.Val(game.PT{Value: 1}),
			Toughness: opt.Val(game.PT{Value: 2}),
			StaticAbilities: []game.StaticAbility{
				game.DeathtouchStaticBody,
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
			Deathtouch
			Revolt — This creature enters with a +1/+1 counter on it if a permanent left the battlefield under your control this turn.
		`,
		},
	}
}
