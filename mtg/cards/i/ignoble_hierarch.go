package i

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// IgnobleHierarch is the card definition for Ignoble Hierarch.
//
// Type: Creature — Goblin Shaman
// Cost: {G}
//
// Oracle text:
//
//	Exalted (Whenever a creature you control attacks alone, that creature gets +1/+1 until end of turn.)
//	{T}: Add {B}, {R}, or {G}.
var IgnobleHierarch = &game.CardDef{
	Name: "Ignoble Hierarch",
	ManaCost: opt.Val(mana.Cost{
		mana.ColoredMana(mana.Green),
	}),
	ManaValue:     1,
	Colors:        []mana.Color{mana.Green},
	ColorIdentity: mana.NewColorIdentity(mana.Black, mana.Green, mana.Red),
	Types:         []types.Card{types.Creature},
	Subtypes:      []types.Sub{types.Sub("Goblin"), types.Shaman},
	Power:         opt.Val(game.PT{Value: 0}),
	Toughness:     opt.Val(game.PT{Value: 1}),
	OracleText:    "Exalted (Whenever a creature you control attacks alone, that creature gets +1/+1 until end of turn.)\n{T}: Add {B}, {R}, or {G}.",
	Abilities: []game.AbilityDef{
		game.ExaltedAbility,
		{
			Kind:          game.ActivatedAbility,
			Text:          "{T}: Add {B}, {R}, or {G}.",
			IsManaAbility: true,
			AdditionalCosts: []game.AdditionalCost{
				{Kind: game.AdditionalCostTap},
			},
			Effects: []game.Effect{
				{
					Type:        game.EffectChoose,
					TargetIndex: game.TargetIndexController,
					Choice: opt.Val(game.ResolutionChoice{
						Kind:   game.ResolutionChoiceColor,
						Prompt: "Choose a color",
						Colors: []mana.Color{mana.Black, mana.Red, mana.Green},
					}),
					LinkID: "ignoble-hierarch-color",
				},
				{
					Type:         game.EffectAddMana,
					Amount:       1,
					TargetIndex:  game.TargetIndexController,
					ChoiceLinkID: "ignoble-hierarch-color",
				},
			},
		},
	},
}
