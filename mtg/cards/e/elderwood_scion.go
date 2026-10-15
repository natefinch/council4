package e

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// ElderwoodScion is the card definition for Elderwood Scion.
//
// Type: Creature — Elemental
// Cost: {3}{G}{W}
//
// Oracle text:
//
//	Trample, lifelink
//	Spells you cast that target this creature cost {2} less to cast.
//	Spells your opponents cast that target this creature cost {2} more to cast.
var ElderwoodScion = newElderwoodScion

func newElderwoodScion() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White, color.Green),
		CardFace: game.CardFace{
			Name: "Elderwood Scion",
			ManaCost: opt.Val(cost.Mana{
				cost.O(3),
				cost.G,
				cost.W,
			}),
			Colors:    []color.Color{color.Green, color.White},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Elemental},
			Power:     opt.Val(game.PT{Value: 4}),
			Toughness: opt.Val(game.PT{Value: 4}),
			StaticAbilities: []game.StaticAbility{
				game.TrampleStaticBody,
				game.LifelinkStaticBody,
				game.StaticAbility{
					Text: "Spells you cast that target this creature cost {2} less to cast.",
					RuleEffects: []game.RuleEffect{
						game.RuleEffect{
							Kind:           game.RuleEffectCostModifier,
							Duration:       game.DurationPermanent,
							AffectedPlayer: game.PlayerYou,
							CostModifier: game.CostModifier{
								Kind:             game.CostModifierSpell,
								GenericReduction: 2,
								TargetsSource:    true,
							},
						},
					},
				},
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
			Trample, lifelink
			Spells you cast that target this creature cost {2} less to cast.
			Spells your opponents cast that target this creature cost {2} more to cast.
		`,
		},
	}
}
