package c

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/opt"
)

// Command Tower
//
// Type: Land
//
// Oracle text:
//
//	{T}: Add one mana of any color in your commander's color identity.
//
// Missing primitives:
//   - ResolutionChoice.Colors is a static slice; it cannot express "the colors in your
//     commander's color identity," which is a dynamic game-state query. The approximation
//     below offers all five colors; ImplementationID "command-tower" must restrict the
//     choice to the controller's commander's color identity at activation time.
var CommandTower = &game.CardDef{
	Name:             "Command Tower",
	ManaValue:        0,
	Types:            []game.CardType{game.TypeLand},
	OracleText:       "{T}: Add one mana of any color in your commander's color identity.",
	ImplementationID: "command-tower",
	Abilities: []game.AbilityDef{
		{
			Kind:          game.ActivatedAbility,
			Text:          "{T}: Add one mana of any color in your commander's color identity.",
			IsManaAbility: true,
			AdditionalCosts: []game.AdditionalCost{
				{Kind: game.AdditionalCostTap},
			},
			Effects: []game.Effect{
				{
					// All five colors listed as approximation; ImplementationID narrows
					// this to the commander's color identity at runtime.
					Type:        game.EffectChoose,
					TargetIndex: -1,
					Choice: opt.Val(game.ResolutionChoice{
						Kind:   game.ResolutionChoiceColor,
						Prompt: "Choose a color in your commander's color identity",
						Colors: []mana.Color{
							mana.White, mana.Blue, mana.Black, mana.Red, mana.Green,
						},
					}),
					LinkID: "command-tower-color",
				},
				{
					Type:         game.EffectAddMana,
					Amount:       1,
					TargetIndex:  -1,
					ChoiceLinkID: "command-tower-color",
				},
			},
		},
	},
}
