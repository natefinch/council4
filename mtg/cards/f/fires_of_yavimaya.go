package f

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// FiresOfYavimaya is the card definition for Fires of Yavimaya.
//
// Type: Enchantment
// Cost: {1}{R}{G}
//
// Oracle text:
//
//	Creatures you control have haste.
//	Sacrifice this enchantment: Target creature gets +2/+2 until end of turn.
var FiresOfYavimaya = &game.CardDef{
	Name: "Fires of Yavimaya",
	ManaCost: opt.Val(mana.Cost{
		mana.GenericMana(1),
		mana.R,
		mana.G,
	}),
	Colors:        []color.Color{color.Green, color.Red},
	ColorIdentity: mana.NewColorIdentity(color.Green, color.Red),
	Types:         []types.Card{types.Enchantment},
	OracleText:    "Creatures you control have haste.\nSacrifice this enchantment: Target creature gets +2/+2 until end of turn.",
	Abilities: []game.AbilityDef{
		{
			Kind: game.StaticAbility,
			Text: "Creatures you control have haste.",
			Effects: []game.Effect{
				{
					Type:        game.EffectApplyContinuous,
					TargetIndex: game.TargetIndexController,
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
		{
			Kind: game.ActivatedAbility,
			Text: "Sacrifice this enchantment: Target creature gets +2/+2 until end of turn.",
			AdditionalCosts: []game.AdditionalCost{
				{Kind: game.AdditionalCostSacrificeSource},
			},
			Targets: []game.TargetSpec{
				{
					MinTargets: 1,
					MaxTargets: 1,
					Constraint: "creature",
					Allow:      game.TargetAllowPermanent,
					Predicate: game.TargetPredicate{
						PermanentTypes: []types.Card{types.Creature},
					},
				},
			},
			Effects: []game.Effect{
				{
					Type:           game.EffectModifyPT,
					TargetIndex:    0,
					PowerDelta:     2,
					ToughnessDelta: 2,
					UntilEndOfTurn: true,
				},
			},
		},
	},
}
