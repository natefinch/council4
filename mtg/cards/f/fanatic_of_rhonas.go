package f

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/opt"
)

// Fanatic of Rhonas
//
// Type: Creature — Snake Druid
// Cost: {1}{G}
//
// Oracle text:
//
//	{T}: Add {G}.
//	Ferocious — {T}: Add {G}{G}{G}{G}. Activate only if you control a creature with power 4 or greater.
//	Eternalize {2}{G}{G}
var FanaticOfRhonas = &game.CardDef{
	Name: "Fanatic of Rhonas",
	ManaCost: opt.Val(mana.Cost{
		mana.GenericMana(1),
		mana.ColoredMana(mana.Green),
	}),
	ManaValue:     2,
	Colors:        []mana.Color{mana.Green},
	ColorIdentity: mana.NewColorIdentity(mana.Green),
	Types:         []game.CardType{game.TypeCreature},
	Subtypes:      []string{game.CreatureSubtypeSnake, game.CreatureSubtypeDruid},
	Power:         opt.Val(game.PT{Value: 1}),
	Toughness:     opt.Val(game.PT{Value: 4}),
	OracleText:    "{T}: Add {G}.\nFerocious — {T}: Add {G}{G}{G}{G}. Activate only if you control a creature with power 4 or greater.\nEternalize {2}{G}{G} ({2}{G}{G}, Exile this card from your graveyard: Create a token that's a copy of it, except it's a 4/4 black Zombie Snake Druid with no mana cost. Eternalize only as a sorcery.)",
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
		{
			Kind:          game.ActivatedAbility,
			Text:          "Ferocious — {T}: Add {G}{G}{G}{G}. Activate only if you control a creature with power 4 or greater.",
			IsManaAbility: true,
			AdditionalCosts: []game.AdditionalCost{
				{Kind: game.AdditionalCostTap},
			},
			ActivationCondition: opt.Val(game.Condition{
				Text: "you control a creature with power 4 or greater",
				ControllerControls: game.PermanentFilter{
					Types: []game.CardType{game.TypeCreature},
					Power: opt.Val(game.IntComparison{
						Op:    game.CompareGreaterOrEqual,
						Value: 4,
					}),
				},
			}),
			Effects: []game.Effect{
				{Type: game.EffectAddMana, Amount: 1, ManaColor: mana.Green, TargetIndex: -1},
				{Type: game.EffectAddMana, Amount: 1, ManaColor: mana.Green, TargetIndex: -1},
				{Type: game.EffectAddMana, Amount: 1, ManaColor: mana.Green, TargetIndex: -1},
				{Type: game.EffectAddMana, Amount: 1, ManaColor: mana.Green, TargetIndex: -1},
			},
		},
		game.EternalizeAbility(
			mana.Cost{mana.GenericMana(2), mana.ColoredMana(mana.Green), mana.ColoredMana(mana.Green)},
			game.CreatureSubtypeSnake, game.CreatureSubtypeDruid,
		),
	},
}
