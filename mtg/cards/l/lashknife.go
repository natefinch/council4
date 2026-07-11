package l

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// Lashknife is the card definition for Lashknife.
//
// Type: Enchantment — Aura
// Cost: {1}{W}
//
// Oracle text:
//
//	If you control a Plains, you may tap an untapped creature you control rather than pay this spell's mana cost.
//	Enchant creature
//	Enchanted creature has first strike.
var Lashknife = newLashknife

func newLashknife() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White),
		CardFace: game.CardFace{
			Name: "Lashknife",
			ManaCost: opt.Val(cost.Mana{
				cost.O(1),
				cost.W,
			}),
			Colors:   []color.Color{color.White},
			Types:    []types.Card{types.Enchantment},
			Subtypes: []types.Sub{types.Aura},
			StaticAbilities: []game.StaticAbility{
				game.EnchantStaticAbility(&game.TargetSpec{
					MinTargets: 1,
					MaxTargets: 1,
					Constraint: "creature",
					Allow:      game.TargetAllowPermanent,
					Selection:  opt.Val(game.Selection{RequiredTypesAny: []types.Card{types.Creature}}),
				}),
				game.StaticAbility{
					ContinuousEffects: []game.ContinuousEffect{
						game.ContinuousEffect{
							Layer: game.LayerAbility,
							Group: game.AttachedObjectGroup(game.SourcePermanentReference()),
							AddKeywords: []game.Keyword{
								game.FirstStrike,
							},
						},
					},
				},
			},
			AlternativeCosts: []cost.Alternative{
				cost.Alternative{
					Label: "Tap an untapped creature you control",
					AdditionalCosts: []cost.Additional{
						{
							Kind:               cost.AdditionalTapPermanents,
							Text:               "tap an untapped creature you control",
							Amount:             1,
							MatchPermanentType: true,
							PermanentType:      types.Creature,
						},
					},
					Condition:        cost.AlternativeConditionControlsPermanentSubtype,
					ConditionSubtype: types.Plains,
				},
			},
			OracleText: `
			If you control a Plains, you may tap an untapped creature you control rather than pay this spell's mana cost.
			Enchant creature
			Enchanted creature has first strike.
		`,
		},
	}
}
