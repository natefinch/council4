package d

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// Deathcap Glade
//
// Type: Land
//
// Oracle text:
//
//	This land enters tapped unless you control two or more other lands.
//	{T}: Add {B} or {G}.
var DeathcapGlade = &game.CardDef{
	Name:          "Deathcap Glade",
	ManaValue:     0,
	ColorIdentity: mana.NewColorIdentity(mana.Black, mana.Green),
	Types:         []types.Card{types.Land},
	OracleText:    "This land enters tapped unless you control two or more other lands.\n{T}: Add {B} or {G}.",
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
			Text:          "{T}: Add {B} or {G}.",
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
						Colors: []mana.Color{mana.Black, mana.Green},
					}),
					LinkID: "deathcap-glade-color",
				},
				{
					Type:         game.EffectAddMana,
					Amount:       1,
					TargetIndex:  -1,
					ChoiceLinkID: "deathcap-glade-color",
				},
			},
		},
	},
}
