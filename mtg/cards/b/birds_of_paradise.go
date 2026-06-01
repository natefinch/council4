package b

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// Birds of Paradise
//
// Type: Creature — Bird
// Cost: {G}
//
// Oracle text:
//
//	Flying
//	{T}: Add one mana of any color.

var BirdsOfParadise = &game.CardDef{
	Name: "Birds of Paradise",
	ManaCost: opt.Val(mana.Cost{
		mana.ColoredMana(mana.Green),
	}),
	ManaValue:     1,
	Colors:        []mana.Color{mana.Green},
	ColorIdentity: mana.NewColorIdentity(mana.Green),
	Types:         []types.Card{types.Creature},
	Subtypes:      []types.Sub{types.Bird},
	Power:         opt.Val(game.PT{Value: 0}),
	Toughness:     opt.Val(game.PT{Value: 1}),
	OracleText:    "Flying\n{T}: Add one mana of any color.",
	Abilities: []game.AbilityDef{
		{
			Kind:     game.StaticAbility,
			Text:     "Flying",
			Keywords: []game.Keyword{game.Flying},
		},
		{
			Kind:          game.ActivatedAbility,
			Text:          "{T}: Add one mana of any color.",
			IsManaAbility: true,
			AdditionalCosts: []game.AdditionalCost{
				{Kind: game.AdditionalCostTap},
			},
			Effects: []game.Effect{
				{
					Type:        game.EffectChoose,
					TargetIndex: -1,
					Choice: opt.Val(game.ResolutionChoice{
						Kind:   game.ResolutionChoiceColor,
						Prompt: "Choose a color",
						Colors: []mana.Color{
							mana.White, mana.Blue, mana.Black, mana.Red, mana.Green,
						},
					}),
					LinkID: "birds-color",
				},
				{
					Type:         game.EffectAddMana,
					Amount:       1,
					TargetIndex:  -1,
					ChoiceLinkID: "birds-color",
				},
			},
		},
	},
}
