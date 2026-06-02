package b

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// BasiliskCollar is the card definition for Basilisk Collar.
//
// Type: Artifact — Equipment
// Cost: {1}
//
// Oracle text:
//
//	Equipped creature has deathtouch and lifelink.
//	Equip {2}
var BasiliskCollar = &game.CardDef{
	Name: "Basilisk Collar",
	ManaCost: opt.Val(cost.Mana{
		cost.O(1),
	}),
	Types:      []types.Card{types.Artifact},
	Subtypes:   []types.Sub{types.Equipment},
	OracleText: "Equipped creature has deathtouch and lifelink. (Any amount of damage it deals to a creature is enough to destroy it. Damage dealt by this creature also causes you to gain that much life.)\nEquip {2} ({2}: Attach to target creature you control. Equip only as a sorcery.)",
	Abilities: []game.AbilityDef{
		{
			Kind: game.StaticAbility,
			Text: "Equipped creature has deathtouch and lifelink.",
			Effects: []game.Effect{
				{
					Type:        game.EffectApplyContinuous,
					TargetIndex: game.TargetIndexSourcePermanent,
					ContinuousEffects: []game.ContinuousEffect{
						{
							Layer:       game.LayerAbility,
							Selector:    game.EffectSelectorEquippedCreature,
							AddKeywords: []game.Keyword{game.Deathtouch, game.Lifelink},
						},
					},
				},
			},
		},
		{
			// EffectAttach (type 27) is not executed by the rules engine; the Equip
			// keyword together with ManaCost, Timing, and Targets is sufficient for
			// the rules layer to perform attachment.
			Kind:     game.ActivatedAbility,
			Text:     "Equip {2}",
			Keywords: []game.Keyword{game.Equip},
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
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
