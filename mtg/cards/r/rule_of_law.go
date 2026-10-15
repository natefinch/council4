package r

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// RuleOfLaw is the card definition for Rule of Law.
//
// Type: Enchantment
// Cost: {2}{W}
//
// Oracle text:
//
//	Each player can't cast more than one spell each turn.
var RuleOfLaw = newRuleOfLaw

func newRuleOfLaw() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White),
		CardFace: game.CardFace{
			Name: "Rule of Law",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
				cost.W,
			}),
			Colors: []color.Color{color.White},
			Types:  []types.Card{types.Enchantment},
			StaticAbilities: []game.StaticAbility{
				game.StaticAbility{
					Text: "Each player can't cast more than one spell each turn.",
					RuleEffects: []game.RuleEffect{
						game.RuleEffect{
							Kind:             game.RuleEffectCastLimitPerTurn,
							Duration:         game.DurationPermanent,
							CastLimitPerTurn: 1,
						},
					},
				},
			},
			OracleText: `
			Each player can't cast more than one spell each turn.
		`,
		},
	}
}
