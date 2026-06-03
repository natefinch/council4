package d

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// DragonskullSummit is the card definition for Dragonskull Summit.
//
// Type: Land
//
// Oracle text:
//
//	This land enters tapped unless you control a Swamp or a Mountain.
//	{T}: Add {B} or {R}.
var DragonskullSummit = &game.CardDef{CardFace: game.CardFace{Name: "Dragonskull Summit",

	Types:      []types.Card{types.Land},
	OracleText: "This land enters tapped unless you control a Swamp or a Mountain.\n{T}: Add {B} or {R}.",
	ReplacementAbilities: []game.ReplacementAbilityDef{
		game.EntersTappedIfReplacement("This land enters tapped unless you control a Swamp or a Mountain.", &game.Condition{
			Negate: true,
			ControllerControls: game.PermanentFilter{
				SubtypesAny: []types.Sub{types.Swamp, types.Mountain},
			},
		}),
	},
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
					TargetIndex: game.TargetIndexController,
					Choice: opt.Val(game.ResolutionChoice{
						Kind:   game.ResolutionChoiceMana,
						Prompt: "Choose a color",
						Colors: []mana.Color{mana.B, mana.R},
					}),
					LinkID: "dragonskull-summit-color",
				},
				{
					Type:         game.EffectAddMana,
					Amount:       1,
					TargetIndex:  game.TargetIndexController,
					ChoiceLinkID: "dragonskull-summit-color",
				},
			},
		},
	}}, ColorIdentity: color.NewIdentity(color.Black, color.Red),
}
