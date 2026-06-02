package c

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// CinderGlade is the card definition for Cinder Glade.
//
// Type: Land — Mountain Forest
//
// Oracle text:
//
//	({T}: Add {R} or {G}.)
//	This land enters tapped unless you control two or more basic lands.
//
// The parenthetical mana ability is reminder text for the Mountain and Forest
// subtypes. It is modelled explicitly because council4 does not auto-derive
// subtype mana abilities at runtime.
var CinderGlade = &game.CardDef{
	Name:          "Cinder Glade",
	ColorIdentity: mana.NewColorIdentity(mana.Green, mana.Red),
	Types:         []types.Card{types.Land},
	Subtypes:      []types.Sub{types.Mountain, types.Forest},
	OracleText:    "({T}: Add {R} or {G}.)\nThis land enters tapped unless you control two or more basic lands.",
	EntersTappedCondition: opt.Val(game.Condition{
		Negate: true,
		ControllerControls: game.PermanentFilter{
			Types:      []types.Card{types.Land},
			Supertypes: []types.Super{types.Basic},
			MinCount:   2,
		},
	}),
	Abilities: []game.AbilityDef{
		{
			Kind:          game.ActivatedAbility,
			Text:          "{T}: Add {R} or {G}.",
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
						Colors: []mana.Color{mana.Red, mana.Green},
					}),
					LinkID: "cinder-glade-color",
				},
				{
					Type:         game.EffectAddMana,
					Amount:       1,
					TargetIndex:  game.TargetIndexController,
					ChoiceLinkID: "cinder-glade-color",
				},
			},
		},
	},
}
