package h

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// HalanaAndAlenaPartners is the card definition for Halana and Alena, Partners.
//
// Type: Legendary Creature — Human Ranger
// Cost: {2}{R}{G}
//
// Oracle text:
//
//	First strike (This creature deals combat damage before creatures without first strike.)
//	Reach (This creature can block creatures with flying.)
//	At the beginning of combat on your turn, put X +1/+1 counters on another target creature you control, where X is Halana and Alena's power. That creature gains haste until end of turn.
var HalanaAndAlenaPartners = &game.CardDef{
	Name: "Halana and Alena, Partners",
	ManaCost: opt.Val(mana.Cost{
		mana.GenericMana(2),
		mana.ColoredMana(mana.Red),
		mana.ColoredMana(mana.Green),
	}),
	ManaValue:     4,
	Colors:        []mana.Color{mana.Green, mana.Red},
	ColorIdentity: mana.NewColorIdentity(mana.Green, mana.Red),
	Supertypes:    []types.Super{types.Legendary},
	Types:         []types.Card{types.Creature},
	Subtypes:      []types.Sub{types.Human, types.Sub("Ranger")},
	Power:         opt.Val(game.PT{Value: 2}),
	Toughness:     opt.Val(game.PT{Value: 3}),
	OracleText:    "First strike (This creature deals combat damage before creatures without first strike.)\nReach (This creature can block creatures with flying.)\nAt the beginning of combat on your turn, put X +1/+1 counters on another target creature you control, where X is Halana and Alena's power. That creature gains haste until end of turn.",
	Abilities: []game.AbilityDef{
		{
			Kind:     game.StaticAbility,
			Text:     "First strike\nReach",
			Keywords: []game.Keyword{game.FirstStrike, game.Reach},
		},
		{
			Kind: game.TriggeredAbility,
			Text: "At the beginning of combat on your turn, put X +1/+1 counters on another target creature you control, where X is Halana and Alena's power. That creature gains haste until end of turn.",
			Trigger: opt.Val(game.TriggerCondition{
				Type: game.TriggerAt,
				Pattern: game.TriggerPattern{
					Event: game.EventBeginningOfStep,
					Step:  game.StepBeginningOfCombat,
				},
			}),
			Targets: []game.TargetSpec{
				{
					MinTargets: 1,
					MaxTargets: 1,
					Constraint: "another creature you control",
					Allow:      game.TargetAllowPermanent,
					Predicate: game.TargetPredicate{
						PermanentTypes: []types.Card{types.Creature},
						Controller:     game.ControllerYou,
						Another:        true,
					},
				},
			},
			Effects: []game.Effect{
				{
					Type:        game.EffectAddCounter,
					TargetIndex: 0,
					CounterKind: counter.PlusOnePlusOne,
					DynamicAmount: opt.Val(game.DynamicAmount{
						Kind:        game.DynamicAmountTargetPower,
						TargetIndex: -2,
					}),
				},
				{
					Type:           game.EffectApplyContinuous,
					TargetIndex:    0,
					UntilEndOfTurn: true,
					ContinuousEffects: []game.ContinuousEffect{
						{
							Layer:       game.LayerAbility,
							AddKeywords: []game.Keyword{game.Haste},
						},
					},
				},
			},
		},
	},
}
