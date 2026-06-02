package k

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// KhalniAmbush is the card definition for Khalni Ambush // Khalni Territory.
//
// Type: Instant // Land
// Face: Khalni Territory — Land
//
// Oracle text:
//
//	Target creature you control fights target creature you don't control. (Each deals damage equal to its power to the other.)
var KhalniAmbush = &game.CardDef{
	Name: "Khalni Ambush // Khalni Territory",
	ManaCost: opt.Val(mana.Cost{
		mana.GenericMana(2),
		mana.ColoredMana(mana.Green),
	}),
	Colors:        []mana.Color{mana.Green},
	ColorIdentity: mana.NewColorIdentity(mana.Green),
	Types:         []types.Card{types.Instant},
	OracleText:    "Target creature you control fights target creature you don't control. (Each deals damage equal to its power to the other.)",
	Abilities: []game.AbilityDef{
		{
			Kind: game.SpellAbility,
			Text: "Target creature you control fights target creature you don't control.",
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
				{
					MinTargets: 1,
					MaxTargets: 1,
					Constraint: "creature you don't control",
					Allow:      game.TargetAllowPermanent,
					Predicate: game.TargetPredicate{
						PermanentTypes: []types.Card{types.Creature},
						Controller:     game.ControllerNotYou,
					},
				},
			},
			Effects: []game.Effect{
				{Type: game.EffectFight},
			},
		},
	},
	Layout: game.LayoutModalDFC,
	Back: opt.Val(game.CardFace{
		Name:         "Khalni Territory",
		Types:        []types.Card{types.Land},
		EntersTapped: true,
		OracleText:   "This land enters tapped.\n{T}: Add {G}.",
		Abilities: []game.AbilityDef{
			{
				Kind:          game.ActivatedAbility,
				Text:          "{T}: Add {G}.",
				IsManaAbility: true,
				AdditionalCosts: []game.AdditionalCost{
					{Kind: game.AdditionalCostTap},
				},
				Effects: []game.Effect{
					{Type: game.EffectAddMana, Amount: 1, ManaColor: mana.Green, TargetIndex: game.TargetIndexController},
				},
			},
		},
	}),
}
