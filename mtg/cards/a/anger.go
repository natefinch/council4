package a

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/opt"
)

// Anger
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
	ManaCost: opt.Val(mana.Cost{
		mana.GenericMana(3),
		mana.ColoredMana(mana.Red),
	}),
	ManaValue:     4,
	Colors:        []mana.Color{mana.Red},
	ColorIdentity: mana.NewColorIdentity(mana.Red),
	Types:         []game.CardType{game.TypeCreature},
	Subtypes:      []string{"Incarnation"},
	Power:         opt.Val(game.PT{Value: 2}),
	Toughness:     opt.Val(game.PT{Value: 2}),
	OracleText:    "Haste\nAs long as this card is in your graveyard and you control a Mountain, creatures you control have haste.",
	Abilities: []game.AbilityDef{
		{
			Kind:     game.StaticAbility,
			Text:     "Haste",
			Keywords: []game.Keyword{game.Haste},
		},
		{
			Kind:           game.StaticAbility,
			Text:           "As long as this card is in your graveyard and you control a Mountain, creatures you control have haste.",
			ZoneOfFunction: game.ZoneGraveyard,
			Condition: opt.Val(game.Condition{
				ControllerControls: game.PermanentFilter{
					SubtypesAny: []string{"Mountain"},
				},
			}),
			Effects: []game.Effect{
				{
					Type:        game.EffectApplyContinuous,
					TargetIndex: -2,
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
