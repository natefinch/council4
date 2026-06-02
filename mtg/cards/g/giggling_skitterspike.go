package g

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// GigglingSkitterspike is the card definition for Giggling Skitterspike.
//
// Type: Artifact Creature — Toy
// Cost: {4}
//
// Oracle text:
//
//	Indestructible
//	Whenever this creature attacks, blocks, or becomes the target of a spell, it deals damage equal to its power to each opponent.
//	{5}: Monstrosity 5. (If this creature isn't monstrous, put five +1/+1 counters on it and it becomes monstrous.)
var GigglingSkitterspike = &game.CardDef{
	Name: "Giggling Skitterspike",
	ManaCost: opt.Val(cost.Mana{
		cost.O(4),
	}),
	Types:      []types.Card{types.Artifact, types.Creature},
	Subtypes:   []types.Sub{types.Toy},
	Power:      opt.Val(game.PT{Value: 1}),
	Toughness:  opt.Val(game.PT{Value: 1}),
	OracleText: "Indestructible\nWhenever this creature attacks, blocks, or becomes the target of a spell, it deals damage equal to its power to each opponent.\n{5}: Monstrosity 5. (If this creature isn't monstrous, put five +1/+1 counters on it and it becomes monstrous.)",
	Abilities: []game.AbilityDef{
		game.IndestructibleAbility,
		{
			Kind: game.TriggeredAbility,
			Text: "Whenever this creature attacks, blocks, or becomes the target of a spell, it deals damage equal to its power to each opponent.",
			Trigger: opt.Val(game.TriggerCondition{
				Type: game.TriggerWhenever,
				Pattern: game.TriggerPattern{
					Event:  game.EventAttackerDeclared,
					Source: game.TriggerSourceSelf,
				},
			}),
			Effects: []game.Effect{
				{
					Type:           game.EffectDamage,
					TargetIndex:    game.TargetIndexController,
					PlayerSelector: game.PlayerSelectorOpponents,
					DynamicAmount: opt.Val(game.DynamicAmount{
						Kind:   game.DynamicAmountObjectPower,
						Object: game.ObjectReference{Kind: game.ObjectReferenceSourcePermanent},
					}),
				},
			},
		},
		{
			Kind: game.TriggeredAbility,
			Text: "Whenever this creature attacks, blocks, or becomes the target of a spell, it deals damage equal to its power to each opponent.",
			Trigger: opt.Val(game.TriggerCondition{
				Type: game.TriggerWhenever,
				Pattern: game.TriggerPattern{
					Event:  game.EventBlockerDeclared,
					Source: game.TriggerSourceSelf,
				},
			}),
			Effects: []game.Effect{
				{
					Type:           game.EffectDamage,
					TargetIndex:    game.TargetIndexController,
					PlayerSelector: game.PlayerSelectorOpponents,
					DynamicAmount: opt.Val(game.DynamicAmount{
						Kind:   game.DynamicAmountObjectPower,
						Object: game.ObjectReference{Kind: game.ObjectReferenceSourcePermanent},
					}),
				},
			},
		},
		{
			Kind: game.TriggeredAbility,
			Text: "Whenever this creature attacks, blocks, or becomes the target of a spell, it deals damage equal to its power to each opponent.",
			Trigger: opt.Val(game.TriggerCondition{
				Type: game.TriggerWhenever,
				Pattern: game.TriggerPattern{
					Event:                game.EventObjectBecameTarget,
					Source:               game.TriggerSourceSelf,
					MatchStackObjectKind: true,
					StackObjectKind:      game.StackSpell,
				},
			}),
			Effects: []game.Effect{
				{
					Type:           game.EffectDamage,
					TargetIndex:    game.TargetIndexController,
					PlayerSelector: game.PlayerSelectorOpponents,
					DynamicAmount: opt.Val(game.DynamicAmount{
						Kind:   game.DynamicAmountObjectPower,
						Object: game.ObjectReference{Kind: game.ObjectReferenceSourcePermanent},
					}),
				},
			},
		},
		{
			Kind: game.ActivatedAbility,
			Text: "{5}: Monstrosity 5. (If this creature isn't monstrous, put five +1/+1 counters on it and it becomes monstrous.)",
			ManaCost: opt.Val(cost.Mana{
				cost.O(5),
			}),
			Timing: game.SorceryOnly,
			Effects: []game.Effect{
				{
					Type:        game.EffectMonstrosity,
					Amount:      5,
					TargetIndex: game.TargetIndexSourcePermanent,
				},
			},
		},
	},
}
