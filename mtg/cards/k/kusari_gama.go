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
var KusariGama = func() *game.CardDef {
	card := &game.CardDef{
		CardFace: game.CardFace{
			Name: "Kusari-Gama",
			ManaCost: opt.Val(cost.Mana{
				cost.O(3),
			}),
			Types:    []types.Card{types.Artifact},
			Subtypes: []types.Sub{types.Equipment},
			OracleText: `
				Equipped creature has "{2}: This creature gets +1/+0 until end of turn."
				Whenever equipped creature deals damage to a blocking creature, this Equipment deals that much damage to each other creature defending player controls.
				Equip {3} ({3}: Attach to target creature you control. Equip only as a sorcery.)
			`,
		},
	}

	card.StaticAbilities = append(card.StaticAbilities, game.StaticAbilityBody{
		Text: `
				Equipped creature has "{2}: This creature gets +1/+0 until end of turn."
			`,
		ContinuousEffects: []game.ContinuousEffect{
			{
				Layer:    game.LayerAbility,
				Selector: game.EffectSelectorEquippedCreature,
				AddAbilities: []game.AbilityBody{
					game.ActivatedAbilityBody{
						Text: `
											{2}: This creature gets +1/+0 until end of turn.
										`,
						ManaCost: opt.Val(cost.Mana{
							cost.O(2),
						}),
						Content: game.Mode{
							Sequence: []game.Instruction{
								{
									Primitive: game.ModifyPT{
										TargetIndex: game.TargetIndexSourcePermanent,
										PowerDelta:  game.Fixed(1),
										Duration:    game.DurationUntilEndOfTurn,
									},
								},
							},
						}.Ability(),
					},
				},
			},
		},
	},
	)

	card.TriggeredAbilities = append(card.TriggeredAbilities,
		game.TriggeredAbilityBody{
			Text: `
				Whenever equipped creature deals damage to a blocking creature, this Equipment deals that much damage to each other creature defending player controls.
			`,
			Trigger: game.TriggerCondition{
				Type: game.TriggerWhenever,
				Pattern: game.TriggerPattern{
					Event:                      game.EventDamageDealt,
					Source:                     game.TriggerSourceAttachedPermanent,
					DamageRecipient:            game.DamageRecipientPermanent,
					DamageRecipientCombatState: game.CombatStateBlocking,
				},
			},
			Content: game.Mode{
				Sequence: []game.Instruction{
					{
						Primitive: game.Damage{
							Amount: game.Dynamic(game.DynamicAmount{
								Kind: game.DynamicAmountEventDamage,
							}),
							Recipient: game.SelectorRecipient(game.EffectSelectorOtherCreaturesDefendingPlayerControls),
						},
					},
				},
			}.Ability(),
		},
	)

	card.ActivatedAbilities = append(card.ActivatedAbilities,
		game.ActivatedAbilityBody{
			Text: `
				Equip {3}
			`,
			ManaCost: opt.Val(cost.Mana{
				cost.O(3),
			}),
			Timing: game.SorceryOnly,
			Content: game.Mode{
				Targets: []game.TargetSpec{
					{
						MinTargets: 1,
						MaxTargets: 1,
						Constraint: "creature you control",
						Allow:      game.TargetAllowPermanent,
						Predicate: game.TargetPredicate{
							PermanentTypes: []types.Card{
								types.Creature,
							},
							Controller: game.ControllerYou,
						},
					},
				},
			}.Ability(),

			KeywordAbilities: []game.KeywordAbility{
				game.EquipKeyword{
					Cost: cost.Mana{
						cost.O(3),
					},
				},
			},
		},
	)
	return card
}()
