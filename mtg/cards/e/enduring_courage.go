package e

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/opt"
)

// Enduring Courage
//
// Type: Enchantment Creature — Dog Glimmer
// Cost: {2}{R}{R}
//
// Oracle text:
//
//	Whenever another creature you control enters, it gets +2/+0 and gains haste until end of turn.
//	When Enduring Courage dies, if it was a creature, return it to the battlefield under its owner's control. It's an enchantment. (It's not a creature.)
var EnduringCourage = &game.CardDef{
	Name: "Enduring Courage",
	ManaCost: opt.Val(mana.Cost{
		mana.GenericMana(2),
		mana.ColoredMana(mana.Red),
		mana.ColoredMana(mana.Red),
	}),
	ManaValue:     4,
	Colors:        []mana.Color{mana.Red},
	ColorIdentity: mana.NewColorIdentity(mana.Red),
	Types:         []game.CardType{game.TypeEnchantment, game.TypeCreature},
	Subtypes:      []string{"Dog", "Glimmer"},
	Power:         opt.Val(game.PT{Value: 3}),
	Toughness:     opt.Val(game.PT{Value: 3}),
	OracleText:    "Whenever another creature you control enters, it gets +2/+0 and gains haste until end of turn.\nWhen Enduring Courage dies, if it was a creature, return it to the battlefield under its owner's control. It's an enchantment. (It's not a creature.)",
	Abilities: []game.AbilityDef{
		{
			Kind: game.TriggeredAbility,
			Text: "Whenever another creature you control enters, it gets +2/+0 and gains haste until end of turn.",
			Trigger: opt.Val(game.TriggerCondition{
				Type: game.TriggerWhenever,
				Pattern: game.TriggerPattern{
					Event:                 game.EventPermanentEnteredBattlefield,
					Controller:            game.TriggerControllerYou,
					ExcludeSelf:           true,
					RequirePermanentTypes: []game.CardType{game.TypeCreature},
				},
			}),
			Effects: []game.Effect{
				{
					Type:           game.EffectModifyPT,
					PowerDelta:     2,
					Object:         opt.Val(game.ObjectReference{Kind: game.ObjectReferenceEventPermanent}),
					UntilEndOfTurn: true,
				},
				{
					Type:   game.EffectApplyContinuous,
					Object: opt.Val(game.ObjectReference{Kind: game.ObjectReferenceEventPermanent}),
					ContinuousEffects: []game.ContinuousEffect{
						{
							Layer:       game.LayerAbility,
							AddKeywords: []game.Keyword{game.Haste},
						},
					},
					UntilEndOfTurn: true,
				},
			},
		},
		{
			Kind: game.TriggeredAbility,
			Text: "When Enduring Courage dies, if it was a creature, return it to the battlefield under its owner's control. It's an enchantment.",
			Trigger: opt.Val(game.TriggerCondition{
				Type: game.TriggerWhen,
				Pattern: game.TriggerPattern{
					Event:  game.EventPermanentDied,
					Source: game.TriggerSourceSelf,
				},
				InterveningIf: "if it was a creature",
				InterveningCondition: opt.Val(game.Condition{
					Text:   "if it was a creature",
					Types:  []game.CardType{game.TypeCreature},
					Object: opt.Val(game.ObjectReference{Kind: game.ObjectReferenceEventPermanent}),
				}),
			}),
			Effects: []game.Effect{
				{
					Type:        game.EffectPutOnBattlefield,
					TargetIndex: -1,
					Card:        opt.Val(game.CardReference{Kind: game.CardReferenceSource}),
					ContinuousEffects: []game.ContinuousEffect{
						{
							Layer:       game.LayerType,
							RemoveTypes: []game.CardType{game.TypeCreature},
						},
					},
				},
			},
		},
	},
}
