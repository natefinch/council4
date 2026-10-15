package w

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// WarNameAspirant is the card definition for War-Name Aspirant.
//
// Type: Creature — Human Warrior
// Cost: {1}{R}
//
// Oracle text:
//
//	Raid — This creature enters with a +1/+1 counter on it if you attacked this turn.
//	This creature can't be blocked by creatures with power 1 or less.
var WarNameAspirant = newWarNameAspirant

func newWarNameAspirant() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Red),
		CardFace: game.CardFace{
			Name: "War-Name Aspirant",
			ManaCost: opt.Val(cost.Mana{
				cost.O(1),
				cost.R,
			}),
			Colors:    []color.Color{color.Red},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Human, types.Warrior},
			Power:     opt.Val(game.PT{Value: 2}),
			Toughness: opt.Val(game.PT{Value: 1}),
			StaticAbilities: []game.StaticAbility{
				game.StaticAbility{
					Text: "This creature can't be blocked by creatures with power 1 or less.",
					RuleEffects: []game.RuleEffect{
						game.RuleEffect{
							Kind:           game.RuleEffectCantBeBlockedByCreaturesWith,
							AffectedSource: true,
							Duration:       game.DurationPermanent,
							BlockerRestriction: game.BlockerRestriction{
								Kind:  game.BlockerRestrictionPowerLessOrEqual,
								Power: 1,
							},
						},
					},
				},
			},
			ReplacementAbilities: []game.ReplacementAbility{
				game.EntersWithCountersIfReplacement("Raid — This creature enters with a +1/+1 counter on it if you attacked this turn.", &game.Condition{
					EventHistory: opt.Val(game.EventHistoryCondition{Pattern: game.TriggerPattern{
						Event:           game.EventAttackerDeclared,
						Controller:      game.TriggerControllerYou,
						StackObjectKind: game.StackSpell,
						CounterKind:     counter.PlusOnePlusOne,
					}, Window: game.EventHistoryCurrentTurn}),
				}, game.CounterPlacement{Kind: counter.PlusOnePlusOne, Amount: 1}),
			},
			OracleText: `
			Raid — This creature enters with a +1/+1 counter on it if you attacked this turn.
			This creature can't be blocked by creatures with power 1 or less.
		`,
		},
	}
}
