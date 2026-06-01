package b

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/opt"
)

// Basilisk Collar
//
// Type: Artifact — Equipment
// Cost: {1}
//
// Oracle text:
//
//	Equipped creature has deathtouch and lifelink.
//	Equip {2}
//
// Missing primitives:
//   - There is no EffectSelectorEquippedCreature (or equivalent) to declaratively
//     target the equipped creature in a static ability. ImplementationID
//     "basilisk-collar" is set on the CardDef so the rules engine can
//     apply the deathtouch/lifelink continuous effect to whatever creature this
//     equipment is currently attached to.
var BasiliskCollar = &game.CardDef{
	Name: "Basilisk Collar",
	ManaCost: opt.Val(mana.Cost{
		mana.GenericMana(1),
	}),
	ManaValue:        1,
	Types:            []game.CardType{game.TypeArtifact},
	Subtypes:         []string{"Equipment"},
	OracleText:       "Equipped creature has deathtouch and lifelink. (Any amount of damage it deals to a creature is enough to destroy it. Damage dealt by this creature also causes you to gain that much life.)\nEquip {2} ({2}: Attach to target creature you control. Equip only as a sorcery.)",
	ImplementationID: "basilisk-collar",
	Abilities: []game.AbilityDef{
		{
			// "Equipped creature has deathtouch and lifelink."
			//
			// No EffectSelectorEquippedCreature exists; Effects are left empty.
			// The CardDef-level ImplementationID "basilisk-collar" delegates to
			// hand-written code that reads Permanent.AttachedTo and applies the
			// layer-6 continuous effects to that creature.
			Kind: game.StaticAbility,
			Text: "Equipped creature has deathtouch and lifelink.",
		},
		{
			// EffectAttach (type 27) is not executed by the rules engine; the Equip
			// keyword together with ManaCost, Timing, and Targets is sufficient for
			// the rules layer to perform attachment.
			Kind:     game.ActivatedAbility,
			Text:     "Equip {2}",
			Keywords: []game.Keyword{game.Equip},
			ManaCost: opt.Val(mana.Cost{
				mana.GenericMana(2),
			}),
			Timing: game.SorceryOnly,
			Targets: []game.TargetSpec{
				{
					MinTargets: 1,
					MaxTargets: 1,
					Constraint: "creature you control",
					Allow:      game.TargetAllowPermanent,
					Predicate: game.TargetPredicate{
						PermanentTypes: []game.CardType{game.TypeCreature},
						Controller:     game.ControllerYou,
					},
				},
			},
		},
	},
}
