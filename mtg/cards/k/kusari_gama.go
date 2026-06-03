package k

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// KusariGama is the card definition for Kusari-Gama.
//
// Type: Artifact — Equipment
// Cost: {3}
//
// Oracle text:
//
//	Equipped creature has "{2}: This creature gets +1/+0 until end of turn."
//	Whenever equipped creature deals damage to a blocking creature, this Equipment deals that much damage to each other creature defending player controls.
//	Equip {3} ({3}: Attach to target creature you control. Equip only as a sorcery.)
var KusariGama = &game.CardDef{CardFace: game.CardFace{Name: "Kusari-Gama",
	ManaCost: opt.Val(cost.Mana{
		cost.O(3),
	}),
	Types:      []types.Card{types.Artifact},
	Subtypes:   []types.Sub{types.Equipment},
	OracleText: "Equipped creature has \"{2}: This creature gets +1/+0 until end of turn.\"\nWhenever equipped creature deals damage to a blocking creature, this Equipment deals that much damage to each other creature defending player controls.\nEquip {3} ({3}: Attach to target creature you control. Equip only as a sorcery.)",
	Abilities: []game.AbilityDef{
		{
			Kind: game.StaticAbility,
			Text: "Equipped creature has \"{2}: This creature gets +1/+0 until end of turn.\"",
			Effects: []game.Effect{
				{
					Type:        game.EffectApplyContinuous,
					TargetIndex: game.TargetIndexSourcePermanent,
					ContinuousEffects: []game.ContinuousEffect{
						{
							Layer:    game.LayerAbility,
							Selector: game.EffectSelectorEquippedCreature,
							AddAbilities: []game.AbilityDef{
								{
									Kind: game.ActivatedAbility,
									Text: "{2}: This creature gets +1/+0 until end of turn.",
									ManaCost: opt.Val(cost.Mana{
										cost.O(2),
									}),
									Effects: []game.Effect{
										{
											Type:           game.EffectModifyPT,
											PowerDelta:     1,
											UntilEndOfTurn: true,
											TargetIndex:    game.TargetIndexSourcePermanent,
										},
									},
								},
							},
						},
					},
				},
			},
		},
		{
			Kind: game.TriggeredAbility,
			Text: "Whenever equipped creature deals damage to a blocking creature, this Equipment deals that much damage to each other creature defending player controls.",
			Trigger: opt.Val(game.TriggerCondition{
				Type: game.TriggerWhenever,
				Pattern: game.TriggerPattern{
					Event:                      game.EventDamageDealt,
					Source:                     game.TriggerSourceAttachedPermanent,
					DamageRecipient:            game.DamageRecipientPermanent,
					DamageRecipientCombatState: game.CombatStateBlocking,
				},
			}),
			Effects: []game.Effect{
				{
					Type:        game.EffectDamage,
					TargetIndex: game.TargetIndexSourcePermanent,
					DynamicAmount: opt.Val(game.DynamicAmount{
						Kind: game.DynamicAmountEventDamage,
					}),
					Selector: game.EffectSelectorOtherCreaturesDefendingPlayerControls,
				},
			},
		},
		{
			Kind:     game.ActivatedAbility,
			Text:     "Equip {3}",
			Keywords: []game.Keyword{game.Equip},
			ManaCost: opt.Val(cost.Mana{
				cost.O(3),
			}),
			Timing: game.SorceryOnly,
			Targets: []game.TargetSpec{
				{
					MinTargets: 1,
					MaxTargets: 1,
					Constraint: "creature you control",
					Allow:      game.TargetAllowPermanent,
					Predicate: game.TargetPredicate{
						PermanentTypes: []types.Card{types.Creature},
						Controller:     game.ControllerYou,
					},
				},
			},
		},
	}},
}
