package b

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/opt"
)

// Blazing Sunsteel
//
// Type: Artifact — Equipment
// Cost: {1}{R}
//
// Oracle text:
//
//	Equipped creature gets +1/+0 for each opponent you have.
//	Whenever equipped creature is dealt damage, it deals that much damage to any target.
//	Equip {4}
//
// Missing primitives:
//   - EffectSelectorEquippedCreature does not exist; the static P/T boost cannot
//     select the equipped creature declaratively.
//   - DynamicAmountCountOpponents does not exist; "for each opponent you have"
//     cannot be expressed as a DynamicAmount.
//   - TriggerPattern has no TriggerSourceEquipped filter; the triggered ability
//     cannot be confined to damage dealt to the equipped creature only.
//   - DynamicAmountEventDamage does not exist; "that much" (the damage amount from
//     the triggering event) cannot be forwarded as a DynamicAmount.
//     ImplementationID "blazing-sunsteel" is set so a hand-written rules handler
//     can apply the continuous +N/+0 effect and wire the damage-redirection trigger.
var BlazingSunsteel = &game.CardDef{
	Name: "Blazing Sunsteel",
	ManaCost: opt.Val(mana.Cost{
		mana.GenericMana(1),
		mana.ColoredMana(mana.Red),
	}),
	ManaValue:        2,
	Colors:           []mana.Color{mana.Red},
	ColorIdentity:    mana.NewColorIdentity(mana.Red),
	Types:            []game.CardType{game.TypeArtifact},
	Subtypes:         []string{"Equipment"},
	OracleText:       "Equipped creature gets +1/+0 for each opponent you have.\nWhenever equipped creature is dealt damage, it deals that much damage to any target.\nEquip {4}",
	ImplementationID: "blazing-sunsteel",
	Abilities: []game.AbilityDef{
		{
			// No EffectSelectorEquippedCreature or DynamicAmountCountOpponents;
			// rules engine applies the +N/+0 effect via ImplementationID.
			Kind: game.StaticAbility,
			Text: "Equipped creature gets +1/+0 for each opponent you have.",
		},
		{
			// No TriggerSourceEquipped and no DynamicAmountEventDamage;
			// rules engine wires the damage-redirection trigger via ImplementationID.
			Kind: game.TriggeredAbility,
			Text: "Whenever equipped creature is dealt damage, it deals that much damage to any target.",
			Targets: []game.TargetSpec{
				{
					MinTargets: 1,
					MaxTargets: 1,
					Constraint: "any target",
					Allow:      game.TargetAllowPermanent | game.TargetAllowPlayer,
				},
			},
		},
		{
			Kind:     game.ActivatedAbility,
			Text:     "Equip {4}",
			Keywords: []game.Keyword{game.Equip},
			ManaCost: opt.Val(mana.Cost{
				mana.GenericMana(4),
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
