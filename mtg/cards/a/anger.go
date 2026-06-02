package a

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// Anger is the card definition for Anger.
//
// Type: Creature — Incarnation
// Cost: {3}{R}
//
// Oracle text:
//
//	Haste
//	As long as this card is in your graveyard and you control a Mountain, creatures you control have haste.
var Anger = &game.CardDef{
	Name: "Anger",
	ManaCost: opt.Val(cost.Mana{
		cost.O(3),
		cost.R,
	}),
	Colors:        []color.Color{color.Red},
	ColorIdentity: mana.NewColorIdentity(color.Red),
	Types:         []types.Card{types.Creature},
	Subtypes:      []types.Sub{types.Incarnation},
	Power:         opt.Val(game.PT{Value: 2}),
	Toughness:     opt.Val(game.PT{Value: 2}),
	OracleText:    "Haste\nAs long as this card is in your graveyard and you control a Mountain, creatures you control have haste.",
	Abilities: []game.AbilityDef{
		game.HasteAbility,
		{
			Kind:           game.StaticAbility,
			Text:           "As long as this card is in your graveyard and you control a Mountain, creatures you control have haste.",
			ZoneOfFunction: game.ZoneGraveyard,
			Condition: opt.Val(game.Condition{
				ControllerControls: game.PermanentFilter{
					SubtypesAny: []types.Sub{types.Mountain},
				},
			}),
			Effects: []game.Effect{
				{
					Type: game.EffectApplyContinuous,
					ContinuousEffects: []game.ContinuousEffect{
						{
							Layer:       game.LayerAbility,
							Selector:    game.EffectSelectorCreaturesYouControl,
							AddKeywords: []game.Keyword{game.Haste},
						},
					},
				},
			},
		},
	},
}
