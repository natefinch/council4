package b

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/opt"
)

// Bridgeworks Battle // Tanglespan Bridgeworks
//
// Type: Sorcery // Land
// Face: Bridgeworks Battle — Sorcery ({2}{G})
// Face: Tanglespan Bridgeworks — Land
//
// Front oracle text:
//
//	Target creature you control gets +2/+2 until end of turn. It fights up to
//	one target creature you don't control. (Each deals damage equal to its power
//	to the other.)
//
// Back oracle text:
//
//	As this land enters, you may pay 3 life. If you don't, it enters tapped.
//	{T}: Add {G}.
//
// Missing primitives (back face):
//   - ReplacementEffect has no "pay life to suppress enters-tapped" pattern;
//     the conditional ETB cannot be expressed declaratively.
//     ImplementationID "tanglespan-bridgeworks" on the back face delegates to a
//     hand-written rules handler that prompts for the life payment on entry.
var BridgeworksBattle = &game.CardDef{
	Name: "Bridgeworks Battle // Tanglespan Bridgeworks",
	ManaCost: opt.Val(mana.Cost{
		mana.GenericMana(2),
		mana.ColoredMana(mana.Green),
	}),
	ManaValue:     3,
	Colors:        []mana.Color{mana.Green},
	ColorIdentity: mana.NewColorIdentity(mana.Green),
	Types:         []game.CardType{game.TypeSorcery},
	// Root OracleText intentionally empty for LayoutModalDFC: oracle text lives on each face.
	Layout:    game.LayoutModalDFC,
	Abilities: []game.AbilityDef{},
	Faces: []game.CardFace{
		{
			Name: "Bridgeworks Battle",
			ManaCost: opt.Val(mana.Cost{
				mana.GenericMana(2),
				mana.ColoredMana(mana.Green),
			}),
			ManaValue:  3,
			Colors:     []mana.Color{mana.Green},
			Types:      []game.CardType{game.TypeSorcery},
			OracleText: "Target creature you control gets +2/+2 until end of turn. It fights up to one target creature you don't control. (Each deals damage equal to its power to the other.)",
			Abilities: []game.AbilityDef{
				{
					Kind: game.SpellAbility,
					Text: "Target creature you control gets +2/+2 until end of turn. It fights up to one target creature you don't control.",
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
						{
							// "up to one" — may choose zero or one
							MinTargets: 0,
							MaxTargets: 1,
							Constraint: "creature you don't control",
							Allow:      game.TargetAllowPermanent,
							Predicate: game.TargetPredicate{
								PermanentTypes: []game.CardType{game.TypeCreature},
								Controller:     game.ControllerNotYou,
							},
						},
					},
					Effects: []game.Effect{
						{
							Type:           game.EffectModifyPT,
							PowerDelta:     2,
							ToughnessDelta: 2,
							TargetIndex:    0,
							UntilEndOfTurn: true,
						},
						{
							Type:        game.EffectFight,
							TargetIndex: 0,
							Description: "target creature you control fights up to one target creature you don't control",
						},
					},
				},
			},
		},
		{
			// Missing primitive: "you may pay 3 life; if you don't, enters tapped"
			// cannot be expressed declaratively. ImplementationID handles the
			// conditional life-payment on entry.
			Name:             "Tanglespan Bridgeworks",
			ManaValue:        0,
			Types:            []game.CardType{game.TypeLand},
			EntersTapped:     true,
			OracleText:       "As this land enters, you may pay 3 life. If you don't, it enters tapped.\n{T}: Add {G}.",
			ImplementationID: "tanglespan-bridgeworks",
			Abilities: []game.AbilityDef{
				{
					Kind:          game.ActivatedAbility,
					Text:          "{T}: Add {G}.",
					IsManaAbility: true,
					AdditionalCosts: []game.AdditionalCost{
						{Kind: game.AdditionalCostTap},
					},
					Effects: []game.Effect{
						{Type: game.EffectAddMana, Amount: 1, ManaColor: mana.Green, TargetIndex: -1},
					},
				},
			},
		},
	},
}
