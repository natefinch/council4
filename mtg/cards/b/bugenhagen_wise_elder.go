package b

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/opt"
)

// Bugenhagen, Wise Elder
//
// Type: Legendary Creature — Human Shaman
// Cost: {1}{G}
//
// Oracle text:
//
//	Reach
//	At the beginning of your upkeep, if you control a creature with power 7 or greater, draw a card.
//	{T}: Add one mana of any color.
var BugenhagenWiseElder = &game.CardDef{
	Name: "Bugenhagen, Wise Elder",
	ManaCost: opt.Val(mana.Cost{
		mana.GenericMana(1),
		mana.ColoredMana(mana.Green),
	}),
	ManaValue:     2,
	Colors:        []mana.Color{mana.Green},
	ColorIdentity: mana.NewColorIdentity(mana.Green),
	Supertypes:    []game.Supertype{game.Legendary},
	Types:         []game.CardType{game.TypeCreature},
	Subtypes:      []string{game.CreatureSubtypeHuman, game.CreatureSubtypeShaman},
	Power:         opt.Val(game.PT{Value: 1}),
	Toughness:     opt.Val(game.PT{Value: 3}),
	OracleText:    "Reach\nAt the beginning of your upkeep, if you control a creature with power 7 or greater, draw a card.\n{T}: Add one mana of any color.",
	Abilities: []game.AbilityDef{
		{
			Kind:     game.StaticAbility,
			Text:     "Reach",
			Keywords: []game.Keyword{game.Reach},
		},
		{
			Kind: game.TriggeredAbility,
			Text: "At the beginning of your upkeep, if you control a creature with power 7 or greater, draw a card.",
			Trigger: opt.Val(game.TriggerCondition{
				Type: game.TriggerAt,
				Pattern: game.TriggerPattern{
					Event:      game.EventBeginningOfStep,
					Controller: game.TriggerControllerYou,
					Step:       game.StepUpkeep,
				},
				InterveningIf: "if you control a creature with power 7 or greater",
				InterveningCondition: opt.Val(game.Condition{
					ControllerControls: game.PermanentFilter{
						Types: []game.CardType{game.TypeCreature},
						Power: opt.Val(game.IntComparison{
							Op:    game.CompareGreaterOrEqual,
							Value: 7,
						}),
					},
				}),
			}),
			Effects: []game.Effect{
				{Type: game.EffectDraw, Amount: 1, TargetIndex: -1},
			},
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
					LinkID: "bugenhagen-color",
				},
				{
					Type:         game.EffectAddMana,
					Amount:       1,
					TargetIndex:  -1,
					ChoiceLinkID: "bugenhagen-color",
				},
			},
		},
	},
}
