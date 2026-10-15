package v

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// VampireOutcasts is the card definition for Vampire Outcasts.
//
// Type: Creature — Vampire
// Cost: {2}{B}{B}
//
// Oracle text:
//
//	Bloodthirst 2 (If an opponent was dealt damage this turn, this creature enters with two +1/+1 counters on it.)
//	Lifelink (Damage dealt by this creature also causes you to gain that much life.)
var VampireOutcasts = newVampireOutcasts

func newVampireOutcasts() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Black),
		CardFace: game.CardFace{
			Name: "Vampire Outcasts",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
				cost.B,
				cost.B,
			}),
			Colors:    []color.Color{color.Black},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Vampire},
			Power:     opt.Val(game.PT{Value: 2}),
			Toughness: opt.Val(game.PT{Value: 2}),
			StaticAbilities: []game.StaticAbility{
				game.LifelinkStaticBody,
			},
			ReplacementAbilities: []game.ReplacementAbility{
				game.EntersWithCountersIfReplacement("Bloodthirst 2", &game.Condition{
					EventHistory: opt.Val(game.EventHistoryCondition{Pattern: game.TriggerPattern{
						Event:           game.EventDamageDealt,
						Player:          game.TriggerPlayerOpponent,
						StackObjectKind: game.StackSpell,
						DamageRecipient: game.DamageRecipientPlayer,
						CounterKind:     counter.PlusOnePlusOne,
					}, Window: game.EventHistoryCurrentTurn}),
				}, game.CounterPlacement{Kind: counter.PlusOnePlusOne, Amount: 2}),
			},
			OracleText: `
			Bloodthirst 2 (If an opponent was dealt damage this turn, this creature enters with two +1/+1 counters on it.)
			Lifelink (Damage dealt by this creature also causes you to gain that much life.)
		`,
		},
	}
}
