package i

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// InfiltrationLens is the card definition for Infiltration Lens.
//
// Type: Artifact — Equipment
// Cost: {1}
//
// Oracle text:
//
//	Whenever equipped creature becomes blocked by a creature, you may draw two cards.
//	Equip {1}
var InfiltrationLens = func() *game.CardDef {
	card := &game.CardDef{
		CardFace: game.CardFace{
			Name: "Infiltration Lens",
			ManaCost: opt.Val(cost.Mana{
				cost.O(1),
			}),
			Types:    []types.Card{types.Artifact},
			Subtypes: []types.Sub{types.Equipment},
			OracleText: `
				Whenever equipped creature becomes blocked by a creature, you may draw two cards.
				Equip {1}
			`,
		},
	}

	card.TriggeredAbilities = append(card.TriggeredAbilities,
		game.TriggeredAbilityBody{
			Text: `
				Whenever equipped creature becomes blocked by a creature, you may draw two cards.
			`,
			Trigger: game.TriggerCondition{
				Type: game.TriggerWhenever,
				Pattern: game.TriggerPattern{
					Event:   game.EventBlockerDeclared,
					Source:  game.TriggerSourceAttachedPermanent,
					Subject: game.TriggerSubjectBlockedAttacker,
					RequirePermanentTypes: []types.Card{
						types.Creature,
					},
				},
			},
			Optional: true,
			Content: game.Mode{
				Sequence: []game.Instruction{
					{
						Primitive: game.Draw{
							Amount:      game.Fixed(2),
							TargetIndex: game.TargetIndexController,
						},
					},
				},
			}.Ability(),
		},
	)

	card.ActivatedAbilities = append(card.ActivatedAbilities,
		game.ActivatedAbilityBody{
			Text: `
				Equip {1}
			`,
			ManaCost: opt.Val(cost.Mana{
				cost.O(1),
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
						cost.O(1),
					},
				},
			},
		},
	)
	return card
}()
