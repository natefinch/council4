package i

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
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
		mana.G,
	}),
	Colors:        []color.Color{color.Green},
	ColorIdentity: mana.NewColorIdentity(color.Black, color.Green, color.Red),
	Types:         []types.Card{types.Creature},
	Subtypes:      []types.Sub{types.Goblin, types.Shaman},
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
						Colors: []color.Color{color.Black, color.Red, color.Green},
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
