package a

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// Arena is the card definition for Arena.
//
// Type: Land
//
// Oracle text:
//
//	{3}, {T}: Tap target creature you control and target creature of an opponent's
//	choice they control. Those creatures fight each other.
var Arena = &game.CardDef{
	CardFace: game.CardFace{
		Name:  "Arena",
		Types: []types.Card{types.Land},
		OracleText: `
			{3}, {T}: Tap target creature you control and target creature of an opponent's choice they control. Those creatures fight each other. (Each deals damage equal to its power to the other.)
		`,
		ActivatedAbilities: []game.ActivatedAbilityBody{
			{
				Text: `
					{3}, {T}: Tap target creature you control and target creature of an opponent's choice they control. Those creatures fight each other.
				`,
				ManaCost: opt.Val(cost.Mana{
					cost.O(3),
				}),
				AdditionalCosts: []game.AdditionalCost{
					{
						Kind: game.AdditionalCostTap,
					},
				},
				Content: game.PlainAbilityContent{
					Targets: []game.TargetSpec{
						{
							MinTargets: 1,
							MaxTargets: 1,
							Constraint: "creature you control",
							Allow:      game.TargetAllowPermanent,
							Predicate: game.TargetPredicate{
								PermanentTypes: []types.Card{
									types.Creature,
								},
								Controller: game.ControllerYou,
							},
						},
						{
							MinTargets: 1,
							MaxTargets: 1,
							Constraint: "creature of an opponent's choice they control",
							Allow:      game.TargetAllowPermanent,
							Predicate: game.TargetPredicate{
								PermanentTypes: []types.Card{
									types.Creature,
								},
								Controller: game.ControllerYou,
							},
							Chooser: game.TargetChooserOpponent,
						},
					},
					Sequence: []game.Instruction{
						{
							Primitive: game.Tap{
								TargetIndex: 0,
							},
						},
						{
							Primitive: game.Tap{
								TargetIndex: 1,
							},
						},
						{
							Primitive: game.Fight{},
						},
					},
				},
			},
		},
	},
}
