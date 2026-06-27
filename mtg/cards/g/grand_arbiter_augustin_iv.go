package g

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// GrandArbiterAugustinIV is the card definition for Grand Arbiter Augustin IV.
//
// Type: Legendary Creature — Human Advisor
// Cost: {2}{W}{U}
//
// Oracle text:
//
//	White spells you cast cost {1} less to cast.
//	Blue spells you cast cost {1} less to cast.
//	Spells your opponents cast cost {1} more to cast.
var GrandArbiterAugustinIV = newGrandArbiterAugustinIV()

func newGrandArbiterAugustinIV() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White, color.Blue),
		CardFace: game.CardFace{
			Name: "Grand Arbiter Augustin IV",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
				cost.W,
				cost.U,
			}),
			Colors:     []color.Color{color.Blue, color.White},
			Supertypes: []types.Super{types.Legendary},
			Types:      []types.Card{types.Creature},
			Subtypes:   []types.Sub{types.Human, types.Advisor},
			Power:      opt.Val(game.PT{Value: 2}),
			Toughness:  opt.Val(game.PT{Value: 3}),
			StaticAbilities: []game.StaticAbility{
				game.StaticAbility{
					RuleEffects: []game.RuleEffect{
						game.RuleEffect{
							Kind:           game.RuleEffectCostModifier,
							AffectedPlayer: game.PlayerYou,
							CostModifier: game.CostModifier{
								Kind:             game.CostModifierSpell,
								CardSelection:    game.Selection{ColorsAny: []color.Color{color.White}},
								GenericReduction: 1,
							},
						},
					},
				},
				game.StaticAbility{
					RuleEffects: []game.RuleEffect{
						game.RuleEffect{
							Kind:           game.RuleEffectCostModifier,
							AffectedPlayer: game.PlayerYou,
							CostModifier: game.CostModifier{
								Kind:             game.CostModifierSpell,
								CardSelection:    game.Selection{ColorsAny: []color.Color{color.Blue}},
								GenericReduction: 1,
							},
						},
					},
				},
				game.StaticAbility{
					RuleEffects: []game.RuleEffect{
						game.RuleEffect{
							Kind:           game.RuleEffectCostModifier,
							AffectedPlayer: game.PlayerOpponent,
							CostModifier: game.CostModifier{
								Kind:            game.CostModifierSpell,
								GenericIncrease: 1,
							},
						},
					},
				},
			},
			OracleText: `
			White spells you cast cost {1} less to cast.
			Blue spells you cast cost {1} less to cast.
			Spells your opponents cast cost {1} more to cast.
		`,
		},
	}
}
