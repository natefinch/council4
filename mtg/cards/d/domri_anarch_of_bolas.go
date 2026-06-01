package d

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/opt"
)

// Domri, Anarch of Bolas
//
// Type: Legendary Planeswalker — Domri
// Cost: {1}{R}{G}
//
// Oracle text:
//
//	Creatures you control get +1/+0.
//	+1: Add {R} or {G}. Creature spells you cast this turn can't be countered.
//	−2: Target creature you control fights target creature you don't control.
var DomriAnarchOfBolas = &game.CardDef{
	Name: "Domri, Anarch of Bolas",
	ManaCost: opt.Val(mana.Cost{
		mana.GenericMana(1),
		mana.ColoredMana(mana.Red),
		mana.ColoredMana(mana.Green),
	}),
	ManaValue:     3,
	Colors:        []mana.Color{mana.Green, mana.Red},
	ColorIdentity: mana.NewColorIdentity(mana.Green, mana.Red),
	Supertypes:    []game.Supertype{game.Legendary},
	Types:         []game.CardType{game.TypePlaneswalker},
	Subtypes:      []string{"Domri"},
	Loyalty:       opt.Val(3),
	OracleText:    "Creatures you control get +1/+0.\n+1: Add {R} or {G}. Creature spells you cast this turn can't be countered.\n−2: Target creature you control fights target creature you don't control.",
	Abilities: []game.AbilityDef{
		{
			Kind: game.StaticAbility,
			Text: "Creatures you control get +1/+0.",
			Effects: []game.Effect{
				{
					Type:        game.EffectModifyPT,
					PowerDelta:  1,
					TargetIndex: -1,
					Selector:    game.EffectSelectorCreaturesYouControl,
				},
			},
		},
		{
			Kind:             game.ActivatedAbility,
			Text:             "+1: Add {R} or {G}. Creature spells you cast this turn can't be countered.",
			IsLoyaltyAbility: true,
			LoyaltyCost:      1,
			Effects: []game.Effect{
				{
					Type:        game.EffectChoose,
					TargetIndex: -1,
					Choice: opt.Val(game.ResolutionChoice{
						Kind:   game.ResolutionChoiceColor,
						Prompt: "Choose {R} or {G}",
						Colors: []mana.Color{mana.Red, mana.Green},
					}),
					LinkID: "domri-color",
				},
				{
					Type:         game.EffectAddMana,
					Amount:       1,
					TargetIndex:  -1,
					ChoiceLinkID: "domri-color",
				},
				{
					Type:        game.EffectApplyRule,
					TargetIndex: -1,
					Duration:    game.DurationThisTurn,
					RuleEffects: []game.RuleEffect{
						{
							Kind:               game.RuleEffectCantBeCountered,
							AffectedController: game.ControllerYou,
							SpellTypes:         []game.CardType{game.TypeCreature},
						},
					},
				},
			},
		},
		{
			Kind:             game.ActivatedAbility,
			Text:             "−2: Target creature you control fights target creature you don't control.",
			IsLoyaltyAbility: true,
			LoyaltyCost:      -2,
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
					MinTargets: 1,
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
				{Type: game.EffectFight},
			},
		},
	},
}
