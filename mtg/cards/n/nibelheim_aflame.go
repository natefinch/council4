package n

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"

	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// NibelheimAflame is the card definition for Nibelheim Aflame.
//
// Type: Sorcery
// Cost: {2}{R}{R}
//
// Oracle text:
//
//	Choose target creature you control. It deals damage equal to its power to each other creature. If this spell was cast from a graveyard, discard your hand and draw four cards.
//	Flashback {5}{R}{R} (You may cast this card from your graveyard for its flashback cost. Then exile it.)
var NibelheimAflame = &game.CardDef{CardFace: game.CardFace{Name: "Nibelheim Aflame",
	ManaCost: opt.Val(cost.Mana{
		cost.O(2),
		cost.R,
		cost.R,
	}),
	Colors: []color.Color{color.Red},

	Types:      []types.Card{types.Sorcery},
	OracleText: "Choose target creature you control. It deals damage equal to its power to each other creature. If this spell was cast from a graveyard, discard your hand and draw four cards.\nFlashback {5}{R}{R} (You may cast this card from your graveyard for its flashback cost. Then exile it.)",
	Abilities: []game.AbilityDef{
		{
			Kind: game.SpellAbility,
			Text: "Choose target creature you control. It deals damage equal to its power to each other creature. If this spell was cast from a graveyard, discard your hand and draw four cards.",
			AlternativeCosts: []game.AlternativeCost{
				{
					Label: "Flashback",
					ManaCost: opt.Val(cost.Mana{
						cost.O(5),
						cost.R,
						cost.R,
					}),
				},
			},
			Targets: []game.TargetSpec{
				{
					MinTargets: 1,
					MaxTargets: 1,
					Constraint: "creature you control",
					Allow:      game.TargetAllowPermanent,
					Predicate: game.TargetPredicate{
						PermanentTypes: []types.Card{types.Creature},
						Controller:     game.ControllerYou,
					},
				},
			},
			Effects: []game.Effect{
				{
					Type:        game.EffectDamage,
					TargetIndex: 0,
					Selector:    game.EffectSelectorAllCreaturesExceptTarget,
					DamageSource: opt.Val(game.ObjectReference{
						Kind:        game.ObjectReferenceTargetPermanent,
						TargetIndex: 0,
					}),
					DynamicAmount: opt.Val(game.DynamicAmount{
						Kind:        game.DynamicAmountTargetPower,
						TargetIndex: 0,
					}),
					Description: "target creature deals damage equal to its power to each other creature",
				},
				{
					Type:        game.EffectDiscard,
					TargetIndex: game.TargetIndexController,
					DynamicAmount: opt.Val(game.DynamicAmount{
						Kind: game.DynamicAmountControllerHandSize,
					}),
					Condition: opt.Val(game.EffectCondition{
						Condition: opt.Val(game.Condition{CastFromZone: opt.Val(game.ZoneGraveyard)}),
					}),
				},
				{
					Type:        game.EffectDraw,
					Amount:      4,
					TargetIndex: game.TargetIndexController,
					Condition: opt.Val(game.EffectCondition{
						Condition: opt.Val(game.Condition{CastFromZone: opt.Val(game.ZoneGraveyard)}),
					}),
				},
			},
		},
		{
			Kind:     game.StaticAbility,
			Text:     "Flashback {5}{R}{R}",
			Keywords: []game.Keyword{game.Flashback},
		},
	}}, ColorIdentity: color.NewIdentity(color.Red),
}
