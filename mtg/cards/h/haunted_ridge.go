package h

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// HauntedRidge is the card definition for Haunted Ridge.
//
// Type: Land
//
// Oracle text:
//
//	This land enters tapped unless you control two or more other lands.
//	{T}: Add {B} or {R}.
var HauntedRidge = &game.CardDef{
	Name:          "Haunted Ridge",
	ManaValue:     0,
	ColorIdentity: mana.NewColorIdentity(mana.Black, mana.Red),
	Types:         []types.Card{types.Land},
	OracleText:    "This land enters tapped unless you control two or more other lands.\n{T}: Add {B} or {R}.",
	EntersTappedCondition: opt.Val(game.Condition{
		Negate: true,
		ControllerControls: game.PermanentFilter{
			Types:    []types.Card{types.Land},
			MinCount: 2,
		},
	}),
	Abilities: []game.AbilityDef{
		{
			Kind:          game.ActivatedAbility,
			Text:          "{T}: Add {B} or {R}.",
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
						Colors: []mana.Color{mana.Black, mana.Red},
					}),
					LinkID: "haunted-ridge-color",
				},
				{
					Type:         game.EffectAddMana,
					Amount:       1,
					TargetIndex:  -1,
					ChoiceLinkID: "haunted-ridge-color",
				},
			},
		},
	},
}
