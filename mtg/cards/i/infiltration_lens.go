package i

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/mana"
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
var InfiltrationLens = &game.CardDef{
	Name: "Infiltration Lens",
	ManaCost: opt.Val(mana.Cost{
		mana.GenericMana(1),
	}),
	Types:      []types.Card{types.Artifact},
	Subtypes:   []types.Sub{types.Equipment},
	OracleText: "Whenever equipped creature becomes blocked by a creature, you may draw two cards.\nEquip {1}",
	Abilities: []game.AbilityDef{
		{
			Kind:     game.TriggeredAbility,
			Text:     "Whenever equipped creature becomes blocked by a creature, you may draw two cards.",
			Optional: true,
			Trigger: opt.Val(game.TriggerCondition{
				Type: game.TriggerWhenever,
				Pattern: game.TriggerPattern{
					Event:                 game.EventBlockerDeclared,
					Source:                game.TriggerSourceAttachedPermanent,
					Subject:               game.TriggerSubjectBlockedAttacker,
					RequirePermanentTypes: []types.Card{types.Creature},
				},
			}),
			Effects: []game.Effect{
				{Type: game.EffectDraw, Amount: 2, TargetIndex: game.TargetIndexController},
			},
		},
		{
			Kind:     game.ActivatedAbility,
			Text:     "Equip {1}",
			Keywords: []game.Keyword{game.Equip},
			ManaCost: opt.Val(mana.Cost{
				mana.GenericMana(1),
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
	},
}
