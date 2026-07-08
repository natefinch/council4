package v

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// VrynWingmare is the card definition for Vryn Wingmare.
//
// Type: Creature — Pegasus
// Cost: {2}{W}
//
// Oracle text:
//
//	Flying
//	Noncreature spells cost {1} more to cast.
var VrynWingmare = newVrynWingmare

func newVrynWingmare() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White),
		CardFace: game.CardFace{
			Name: "Vryn Wingmare",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
				cost.W,
			}),
			Colors:    []color.Color{color.White},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Pegasus},
			Power:     opt.Val(game.PT{Value: 2}),
			Toughness: opt.Val(game.PT{Value: 1}),
			StaticAbilities: []game.StaticAbility{
				game.FlyingStaticBody,
				game.StaticAbility{
					RuleEffects: []game.RuleEffect{
						game.RuleEffect{
							Kind: game.RuleEffectCostModifier,
							CostModifier: game.CostModifier{
								Kind:            game.CostModifierSpell,
								CardSelection:   game.Selection{ExcludedTypes: []types.Card{types.Creature}},
								GenericIncrease: 1,
							},
						},
					},
				},
			},
			OracleText: `
			Flying
			Noncreature spells cost {1} more to cast.
		`,
		},
	}
}
