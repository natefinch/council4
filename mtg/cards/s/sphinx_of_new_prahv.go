package s

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// SphinxOfNewPrahv is the card definition for Sphinx of New Prahv.
//
// Type: Creature — Sphinx
// Cost: {W}{W}{U}{U}
//
// Oracle text:
//
//	Flying, vigilance
//	Spells your opponents cast that target this creature cost {2} more to cast.
var SphinxOfNewPrahv = newSphinxOfNewPrahv

func newSphinxOfNewPrahv() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White, color.Blue),
		CardFace: game.CardFace{
			Name: "Sphinx of New Prahv",
			ManaCost: opt.Val(cost.Mana{
				cost.W,
				cost.W,
				cost.U,
				cost.U,
			}),
			Colors:    []color.Color{color.Blue, color.White},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Sphinx},
			Power:     opt.Val(game.PT{Value: 4}),
			Toughness: opt.Val(game.PT{Value: 3}),
			StaticAbilities: []game.StaticAbility{
				game.FlyingStaticBody,
				game.VigilanceStaticBody,
				game.StaticAbility{
					Text: "Spells your opponents cast that target this creature cost {2} more to cast.",
					RuleEffects: []game.RuleEffect{
						game.RuleEffect{
							Kind:           game.RuleEffectCostModifier,
							Duration:       game.DurationPermanent,
							AffectedPlayer: game.PlayerOpponent,
							CostModifier: game.CostModifier{
								Kind:            game.CostModifierSpell,
								GenericIncrease: 2,
								TargetsSource:   true,
							},
						},
					},
				},
			},
			OracleText: `
			Flying, vigilance
			Spells your opponents cast that target this creature cost {2} more to cast.
		`,
		},
	}
}
