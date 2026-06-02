package l

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// LegolasMasterArcher is the card definition for Legolas, Master Archer.
//
// Type: Legendary Creature — Elf Archer
// Cost: {1}{G}{G}
//
// Oracle text:
//
//	Reach
//	Whenever you cast a spell that targets Legolas, put a +1/+1 counter on Legolas.
//	Whenever you cast a spell that targets a creature you don't control, Legolas deals damage equal to its power to up to one target creature.
var LegolasMasterArcher = &game.CardDef{
	Name: "Legolas, Master Archer",
	ManaCost: opt.Val(mana.Cost{
		mana.GenericMana(1),
		mana.ColoredMana(mana.Green),
		mana.ColoredMana(mana.Green),
	}),
	Colors:        []mana.Color{mana.Green},
	ColorIdentity: mana.NewColorIdentity(mana.Green),
	Supertypes:    []types.Super{types.Legendary},
	Types:         []types.Card{types.Creature},
	Subtypes:      []types.Sub{types.Sub("Elf"), types.Sub("Archer")},
	Power:         opt.Val(game.PT{Value: 1}),
	Toughness:     opt.Val(game.PT{Value: 4}),
	OracleText:    "Reach\nWhenever you cast a spell that targets Legolas, put a +1/+1 counter on Legolas.\nWhenever you cast a spell that targets a creature you don't control, Legolas deals damage equal to its power to up to one target creature.",
	Abilities: []game.AbilityDef{
		game.ReachAbility,
		{
			Kind: game.TriggeredAbility,
			Text: "Whenever you cast a spell that targets Legolas, put a +1/+1 counter on Legolas.",
			Trigger: opt.Val(game.TriggerCondition{
				Type: game.TriggerWhenever,
				Pattern: game.TriggerPattern{
					Event:              game.EventSpellCast,
					Controller:         game.TriggerControllerYou,
					SpellTargetsSource: true,
				},
			}),
			Effects: []game.Effect{
				{
					Type:        game.EffectAddCounter,
					Amount:      1,
					CounterKind: counter.PlusOnePlusOne,
					TargetIndex: game.TargetIndexSourcePermanent,
				},
			},
		},
		{
			Kind:     game.TriggeredAbility,
			Text:     "Whenever you cast a spell that targets a creature you don't control, Legolas deals damage equal to its power to up to one target creature.",
			Optional: true,
			Trigger: opt.Val(game.TriggerCondition{
				Type: game.TriggerWhenever,
				Pattern: game.TriggerPattern{
					Event:            game.EventSpellCast,
					Controller:       game.TriggerControllerYou,
					SpellTargetAllow: game.TargetAllowPermanent,
					SpellTargetPattern: opt.Val(game.TargetPredicate{
						PermanentTypes: []types.Card{types.Creature},
						Controller:     game.ControllerNotYou,
					}),
				},
			}),
			Targets: []game.TargetSpec{
				{
					MinTargets: 0,
					MaxTargets: 1,
					Constraint: "creature",
					Allow:      game.TargetAllowPermanent,
					Predicate: game.TargetPredicate{
						PermanentTypes: []types.Card{types.Creature},
					},
				},
			},
			Effects: []game.Effect{
				{
					Type:        game.EffectDamage,
					TargetIndex: 0,
					DynamicAmount: opt.Val(game.DynamicAmount{
						Kind:   game.DynamicAmountObjectPower,
						Object: game.ObjectReference{Kind: game.ObjectReferenceSourcePermanent},
					}),
				},
			},
		},
	},
}
