package l

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// LoxodonWarhammer is the card definition for Loxodon Warhammer.
//
// Type: Artifact — Equipment
// Cost: {3}
//
// Oracle text:
//
//	Equipped creature gets +3/+0 and has trample and lifelink.
//	Equip {3}
var LoxodonWarhammer = &game.CardDef{
	Name: "Loxodon Warhammer",
	ManaCost: opt.Val(mana.Cost{
		mana.GenericMana(3),
	}),
	Types:      []types.Card{types.Artifact},
	Subtypes:   []types.Sub{types.Equipment},
	OracleText: "Equipped creature gets +3/+0 and has trample and lifelink.\nEquip {3}",
	Abilities: []game.AbilityDef{
		{
			Kind: game.StaticAbility,
			Text: "Equipped creature gets +3/+0 and has trample and lifelink.",
			Effects: []game.Effect{
				{
					Type:        game.EffectApplyContinuous,
					TargetIndex: game.TargetIndexSourcePermanent,
					ContinuousEffects: []game.ContinuousEffect{
						{
							Layer:          game.LayerPowerToughnessModify,
							Selector:       game.EffectSelectorEquippedCreature,
							PowerDelta:     3,
							ToughnessDelta: 0,
						},
						{
							Layer:       game.LayerAbility,
							Selector:    game.EffectSelectorEquippedCreature,
							AddKeywords: []game.Keyword{game.Trample, game.Lifelink},
						},
					},
				},
			},
		},
		{
			Kind:     game.ActivatedAbility,
			Text:     "Equip {3}",
			Keywords: []game.Keyword{game.Equip},
			ManaCost: opt.Val(mana.Cost{
				mana.GenericMana(3),
			}),
			Timing: game.SorceryOnly,
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
		},
	},
}
