package r

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// RubyLeech is the card definition for Ruby Leech.
//
// Type: Creature — Leech
// Cost: {1}{R}
//
// Oracle text:
//
//	First strike
//	Red spells you cast cost {R} more to cast.
var RubyLeech = newRubyLeech

func newRubyLeech() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Red),
		CardFace: game.CardFace{
			Name: "Ruby Leech",
			ManaCost: opt.Val(cost.Mana{
				cost.O(1),
				cost.R,
			}),
			Colors:    []color.Color{color.Red},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Leech},
			Power:     opt.Val(game.PT{Value: 2}),
			Toughness: opt.Val(game.PT{Value: 2}),
			StaticAbilities: []game.StaticAbility{
				game.FirstStrikeStaticBody,
				game.StaticAbility{
					RuleEffects: []game.RuleEffect{
						game.RuleEffect{
							Kind:           game.RuleEffectCostModifier,
							AffectedPlayer: game.PlayerYou,
							CostModifier: game.CostModifier{
								Kind:            game.CostModifierSpell,
								CardSelection:   game.Selection{ColorsAny: []color.Color{color.Red}},
								ColoredIncrease: []mana.Color{mana.R},
							},
						},
					},
				},
			},
			OracleText: `
			First strike
			Red spells you cast cost {R} more to cast.
		`,
		},
	}
}
